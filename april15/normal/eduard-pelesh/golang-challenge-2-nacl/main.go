package main

import (
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"golang.org/x/crypto/nacl/box"
	"io"
	"log"
	"net"
	"os"
)

type SecureReader struct {
	r   io.Reader
	key [32]byte
}

type SecureWriter struct {
	w   io.Writer
	key [32]byte
}

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	s := new(SecureReader)
	s.r = r
	box.Precompute(&s.key, pub, priv)
	return s
}

func (s *SecureReader) Read(p []byte) (n int, err error) {
	var nonce [24]byte

	if len(p) == 0 {
		return 0, nil
	}

	buf := make([]byte, 32*1024+24)
	m, err := s.r.Read(buf)
	if err != nil && err != io.EOF && m < 24 {
		return 0, err
	}

	copy(nonce[:], buf[:24])

	msg, ok := box.OpenAfterPrecomputation([]byte{}, buf[24:m], &nonce, &s.key)
	if !ok {
		return 0, errors.New("Failed to decrypt a message")
	}

	copy(p[:len(msg)], msg)
	n = len(msg)

	return
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	s := new(SecureWriter)
	s.w = w
	box.Precompute(&s.key, pub, priv)
	return s
}

func (s *SecureWriter) Write(p []byte) (n int, err error) {
	var nonce [24]byte

	if len(p) == 0 {
		return 0, nil
	}

	if _, err := rand.Read(nonce[:]); err != nil {
		return 0, err
	}

	ciphermsg := box.SealAfterPrecomputation(nonce[:], p, &nonce, &s.key)
	n, err = s.w.Write(ciphermsg)

	return
}

type SecureConn struct {
	r io.Reader
	w io.Writer
	c net.Conn
}

func (s *SecureConn) Read(p []byte) (n int, err error) {
	return s.r.Read(p)
}

func (s *SecureConn) Write(p []byte) (n int, err error) {
	return s.w.Write(p)
}

func (s *SecureConn) Close() error {
	return s.c.Close()
}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	c, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	n, err := c.Write(pub[:])
	if err != nil || n != 32 {
		return nil, errors.New("Handshake failed: client key")
	}

	buf := make([]byte, 32)
	m, err := c.Read(buf)
	if (err != nil && err != io.EOF) || m != 32 {
		return nil, errors.New("Handshake failed: server key")
	}

	peer := &[32]byte{}
	copy(peer[:], buf[:])

	s := &SecureConn{
		r: NewSecureReader(c, priv, peer),
		w: NewSecureWriter(c, priv, peer),
		c: c,
	}

	return s, err
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}

	for {
		// Wait for a connection.
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		// Handle the connection in a new goroutine.
		// The loop then returns to accepting, so that
		// multiple connections may be served concurrently.
		go handle(conn, priv, pub)
	}

	return errors.New("Server stopped")
}

// Handle responds to the handshake from client
// and returns an echo.
func handle(c net.Conn, priv, pub *[32]byte) {
	buf := make([]byte, 32)
	m, err := c.Read(buf)
	if (err != nil && err != io.EOF) || m != 32 {
		c.Close()
	}

	peer := &[32]byte{}
	copy(peer[:], buf[:])

	n, err := c.Write(pub[:])
	if err != nil || n != 32 {
		c.Close()
	}

	s := &SecureConn{
		r: NewSecureReader(c, priv, peer),
		w: NewSecureWriter(c, priv, peer),
		c: c,
	}

	io.Copy(s, s)

	// Shut down the connection.
	s.Close()
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
