package main

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"os"
	"sync"

	"golang.org/x/crypto/nacl/box"
)

var (
	// ErrCannotDecrypt is returned by reader if received data cannot be
	// decrypted
	ErrCannotDecrypt = errors.New("cannot decrypt")

	// byte order for size encoding
	endian = binary.BigEndian
)

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	return &secureReader{
		pub:  pub,
		priv: priv,
		r:    r,
	}
}

type secureReader struct {
	pub  *[32]byte
	priv *[32]byte

	mu  sync.Mutex
	r   io.Reader
	buf *bytes.Buffer
}

// Read message, extracting it from encrypted frame.
//
// Every message frame should start with 28 bytes header containing 24 bytes
// nonce, and 4 bytes of uint32 encrypted message length.
func (sr *secureReader) Read(b []byte) (msgLen int, err error) {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	if sr.buf != nil {
		n, err := sr.buf.Read(b)
		if err != io.EOF {
			return n, err
		}
	}

	// there is no more data in the buffer - read new message frame and create
	// new buffer
	var header = make([]byte, 28) // nonce + message size
	if _, err := io.ReadFull(sr.r, header); err != nil {
		return 0, err
	}

	var nonce [24]byte
	copy(nonce[:], header[:24])

	msgSize := endian.Uint32(header[24:])
	encr := make([]byte, msgSize)

	if _, err := io.ReadFull(sr.r, encr); err != nil {
		return 0, err
	}

	raw := make([]byte, 0, len(encr))
	raw, ok := box.Open(raw, encr, &nonce, sr.pub, sr.priv)
	if !ok {
		return 0, ErrCannotDecrypt
	}
	sr.buf = bytes.NewBuffer(raw)
	return sr.buf.Read(b)
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	return &secureWriter{
		pub:  pub,
		priv: priv,
		w:    w,
	}
}

type secureWriter struct {
	pub  *[32]byte
	priv *[32]byte

	mu sync.Mutex
	w  io.Writer
}

// Write given data as encrypted frame.
//
// Every message frame should start with 28 bytes header containing 24 bytes
// nonce, and 4 bytes of uint32 encrypted message length.
func (sw *secureWriter) Write(b []byte) (written int, err error) {
	if len(b) > math.MaxUint32 {
		return 0, errors.New("message too long")
	}

	var nonce [24]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return 0, err
	}
	encr := make([]byte, 0, len(b))
	encr = box.Seal(encr, b, &nonce, sw.pub, sw.priv)

	sw.mu.Lock()
	defer sw.mu.Unlock()

	n, err := sw.w.Write(nonce[:])
	written += n
	if err != nil {
		return written, err
	}

	if err := binary.Write(sw.w, endian, uint32(len(encr))); err != nil {
		return written, err
	}
	written += 4 // uint32 size

	n, err = sw.w.Write(encr)
	written += n
	return written, err
}

// secureConn implementes io.ReadWriteCloser interface, but data exchange is
// encrypted using NaCl
type secureConn struct {
	pub  *[32]byte // their public key
	priv *[32]byte // our private key
	conn net.Conn
	r    io.Reader
	w    io.Writer
}

func (c *secureConn) Read(b []byte) (n int, err error) {
	return c.r.Read(b)
}

func (c *secureConn) Write(b []byte) (n int, err error) {
	return c.w.Write(b)
}

func (c *secureConn) Close() error {
	return c.conn.Close()
}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	c := &secureConn{
		conn: conn,
		pub:  new([32]byte),
		priv: priv,
	}

	// handshake
	// send our public key
	if _, err := c.conn.Write(pub[:]); err != nil {
		return nil, err
	}
	// receive and store their public key
	if _, err := io.ReadFull(c.conn, c.pub[:]); err != nil {
		return nil, err
	}

	c.r = NewSecureReader(conn, c.priv, c.pub)
	c.w = NewSecureWriter(conn, c.priv, c.pub)

	return c, nil
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	for {
		c, err := l.Accept()
		if err != nil {
			return fmt.Errorf("cannot accept connection: %s", err)
		}
		if err := echoConn(c); err != nil {
			return err
		}
	}
}

// echoConn generates private/public key pair and spawns new process to echo
// back all messages received using secure, NaCL encrypted message frames.
//
// For every connection, unique pair of keys is used.
func echoConn(c net.Conn) error {
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		c.Close()
		return err
	}

	go func() {
		defer c.Close()

		// handshake
		// send our public key
		if _, err := c.Write(pub[:]); err != nil {
			return
		}
		// receive client public key
		// we don't need our public key anymore
		if _, err := io.ReadFull(c, pub[:]); err != nil {
			return
		}

		r := NewSecureReader(c, priv, pub)
		w := NewSecureWriter(c, priv, pub)
		io.Copy(w, r)
	}()

	return nil
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
