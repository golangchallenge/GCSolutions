package main

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"golang.org/x/crypto/nacl/box"
)

// MaxSize is the maximum size of a box for transmission.
const MaxSize = 65536 // 16 bits

// SecureReader is a reader implementing NaCl.
type SecureReader struct {
	rd        io.Reader
	ind, size uint16
	buf       []byte
	key       *[32]byte
}

func (sr *SecureReader) Read(p []byte) (n int, err error) {
	if sr.ind >= sr.size {
		var (
			rawSize uint16
			ok      bool
		)
		// Read new box.
		if err = binary.Read(sr.rd, binary.BigEndian, &rawSize); err != nil {
			return
		}
		if rawSize < box.Overhead {
			err = errors.New("box is smaller than box overhead")
			return
		}
		raw := make([]byte, rawSize)
		nonce := &[24]byte{}
		if err = binary.Read(sr.rd, binary.BigEndian, nonce); err != nil {
			return
		}
		if _, err = io.ReadFull(sr.rd, raw); err != nil {
			return
		}
		size := rawSize - box.Overhead
		sr.buf = make([]byte, size)
		if sr.buf, ok = box.OpenAfterPrecomputation(nil, raw, nonce, sr.key); !ok {
			err = errors.New("could not open box")
			return
		}
		sr.ind = 0
		sr.size = size
	}
	n = int(sr.size - sr.ind)
	if l := len(p); l < n {
		n = l
	}
	copy(p, sr.buf[sr.ind:sr.ind+uint16(n)])
	sr.ind += uint16(n)
	return
}

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	sr := &SecureReader{
		rd:  r,
		key: &[32]byte{},
	}
	box.Precompute(sr.key, pub, priv)
	return sr
}

// SecureWriter is a writer that writes using NaCl.
type SecureWriter struct {
	wr  io.Writer
	key *[32]byte
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	sw := &SecureWriter{
		wr:  w,
		key: &[32]byte{},
	}
	box.Precompute(sw.key, pub, priv)
	return sw
}

func (sw *SecureWriter) Write(p []byte) (n int, err error) {
	rawSize := uint16(len(p) + box.Overhead)
	if err = binary.Write(sw.wr, binary.BigEndian, rawSize); err != nil {
		return
	}
	var nonce *[24]byte
	if nonce, err = Nonce(); err != nil {
		return
	}
	if err = binary.Write(sw.wr, binary.BigEndian, nonce); err != nil {
		return
	}
	raw := box.SealAfterPrecomputation(nil, p, nonce, sw.key)
	n, err = sw.wr.Write(raw)
	if n > box.Overhead {
		n -= box.Overhead
	}
	return
}

// Conn is a convenience wrapper for a net.Conn and a secure reader and writer.
type Conn struct {
	io.Reader
	io.Writer
	conn net.Conn
}

// NewConn creates a new instance of Conn wrapping the passed connection.
func NewConn(conn net.Conn) (*Conn, error) {
	pub := &[32]byte{}
	if _, err := io.ReadFull(conn, pub[:]); err != nil {
		return nil, fmt.Errorf("failed to receive server public key: %s", err)
	}
	selfPub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key: %s", err)
	}
	if _, err := conn.Write(selfPub[:]); err != nil {
		return nil, fmt.Errorf("failed to write public key: %s", err)
	}
	return &Conn{
		NewSecureReader(conn, priv, pub),
		NewSecureWriter(conn, priv, pub),
		conn,
	}, nil
}

// Close closes the internal connection.
func (c *Conn) Close() error {
	return c.conn.Close()
}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %s", addr, err)
	}
	return NewConn(conn)
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go func() {
			if err := HandleConn(conn); err != nil {
				log.Printf("failed to handle connection: %s", err)
			}
		}()
	}
}

// HandleConn handles an inbound connection, usually from a listener.
func HandleConn(conn net.Conn) error {
	selfPub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return fmt.Errorf("failed to generate key: %s", err)
	}
	if _, err := conn.Write(selfPub[:]); err != nil {
		return fmt.Errorf("failed to write public key: %s", err)
	}
	pub := &[32]byte{}
	if _, err := io.ReadFull(conn, pub[:]); err != nil {
		return fmt.Errorf("failed to read client public key: %s", err)
	}
	buf := make([]byte, MaxSize)
	sr := NewSecureReader(conn, priv, pub)
	sw := NewSecureWriter(conn, priv, pub)
	for {
		n, err := sr.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to read box: %s", err)
		}
		if _, err := sw.Write(buf[:n]); err != nil {
			return fmt.Errorf("failed to echo back to client: %s", err)
		}
	}
	return nil
}

// Nonce returns a random 24 byte nonce for use with sealing and opening.
func Nonce() (*[24]byte, error) {
	n := &[24]byte{}
	_, err := rand.Read(n[:])
	return n, err
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
	buf := make([]byte, len(os.Args[2]))
	n, err := conn.Read(buf)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", buf[:n])
}
