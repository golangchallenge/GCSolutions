// http://golang-challenge.com/go-challenge2/
package main

import (
	"crypto/rand"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"os"

	"golang.org/x/crypto/nacl/box"
)

// secureReader decrypts messages for an io.Reader object.
type secureReader struct {
	reader    io.Reader
	priv, pub [32]byte
	message   []byte
	r, w      int
}

// NewSecureReader instantiates a new secure reader.
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	return &secureReader{reader: r, priv: *priv, pub: *pub}
}

// Read implements the io.Reader interface.
func (sr *secureReader) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return
	}

	if sr.r == sr.w {
		if err = sr.fill(); err != nil {
			return
		}
	}

	// copy message into p
	n = copy(p, sr.message[sr.r:])
	sr.r += n

	return
}

func (sr *secureReader) fill() (err error) {
	sr.r, sr.w = 0, 0
	sr.message = make([]byte, 0)

	// message struct
	var size [4]byte
	var nonce [24]byte
	var body []byte

	// size
	_, err = io.ReadFull(sr.reader, size[:])
	if err != nil {
		return
	}

	// assert: 0 < n <= math.MaxUint16
	expected := binary.BigEndian.Uint32(size[:])
	if expected == 0 || expected > math.MaxUint16 {
		return fmt.Errorf("unexpected message")
	}

	// nonce
	_, err = io.ReadFull(sr.reader, nonce[:])
	if err != nil {
		return
	}

	// body
	body = make([]byte, expected)
	_, err = io.ReadFull(sr.reader, body)
	if err != nil {
		return
	}

	// open body
	message, ok := box.Open(nil, body, &nonce, &sr.pub, &sr.priv)
	if !ok {
		return fmt.Errorf("unable to open body")
	}

	sr.message = message
	sr.w = len(message)
	return
}

// secureWriter encrypts messages for an io.Writer object.
type secureWriter struct {
	writer    io.Writer
	priv, pub [32]byte
}

// NewSecureWriter instantiates a new secure writer.
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	return &secureWriter{w, *priv, *pub}
}

// Write implements the io.Writer interface.
func (sw *secureWriter) Write(p []byte) (int, error) {
	n := len(p)

	// assert: 0 < n <= math.MaxUint16
	if n == 0 {
		return 0, nil
	}
	if n > math.MaxUint16 {
		return 0, fmt.Errorf("message too large: %d > %d", n, math.MaxUint16)
	}

	// message struct
	var size [4]byte
	var nonce [24]byte
	var body []byte

	// size
	binary.BigEndian.PutUint32(size[:], uint32(box.Overhead+n))
	_, err := sw.writer.Write(size[:])
	if err != nil {
		return 0, err
	}

	// create nonce
	_, err = io.ReadFull(rand.Reader, nonce[:])
	if err != nil {
		return 0, err
	}

	_, err = sw.writer.Write(nonce[:])
	if err != nil {
		return 0, err
	}

	// seal p into body
	body = box.Seal(nil, p, &nonce, &sw.pub, &sw.priv)
	_, err = sw.writer.Write(body)
	if err != nil {
		return 0, err
	}

	return n, err
}

// secureConn provides secured I/O for a net.Conn object.
type secureConn struct {
	net.Conn
	reader io.Reader
	writer io.Writer
}

// handshake generates a private/public key pair,
// perform the handshake and return a secureConn object.
func handshake(conn net.Conn) (c *secureConn, err error) {
	// generate key pair
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return
	}

	// write pubic key
	_, err = conn.Write(pub[:])
	if err != nil {
		return
	}

	// read peer pubic key
	_, err = io.ReadFull(conn, pub[:])
	if err != nil {
		return
	}

	reader := NewSecureReader(conn, priv, pub)
	writer := NewSecureWriter(conn, priv, pub)

	c = &secureConn{conn, reader, writer}
	return
}

// Read implements the io.Reader interface.
func (c *secureConn) Read(p []byte) (n int, err error) {
	return c.reader.Read(p)
}

// Write implements the io.Writer interface.
func (c *secureConn) Write(p []byte) (n int, err error) {
	return c.writer.Write(p)
}

// Dial connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	c, err := handshake(conn)
	if err != nil {
		conn.Close()
	}

	return c, err
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	for {
		// accept incomming connection
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go func(c net.Conn) {
			if err := echo(c); err != nil {
				log.Println("challenge2:", err)
			}
		}(conn)
	}
}

func echo(conn net.Conn) (err error) {
	defer conn.Close()

	c, err := handshake(conn)
	if err != nil {
		return
	}

	// read message
	buf := make([]byte, math.MaxUint16)
	n, err := c.Read(buf)
	if err != nil {
		return
	}

	// echo message
	_, err = c.Write(buf[:n])
	return
}

func main() {
	port := flag.Int("l", 0, "Listen mode. Specify port")
	flag.Parse()

	// Server mode
	if *port != 0 {
		l, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
		if err != nil {
			log.Fatal(err)
		}
		defer l.Close()
		log.Fatal(Serve(l))
	}

	// Client mode
	if len(os.Args) != 3 {
		log.Fatalf("Usage: %s <port> <message>", os.Args[0])
	}
	conn, err := Dial("localhost:" + os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	if _, err := conn.Write([]byte(os.Args[2])); err != nil {
		log.Fatal(err)
	}
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err == io.EOF {
			fmt.Println()
			break
		} else if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s", buf[:n])
	}
}
