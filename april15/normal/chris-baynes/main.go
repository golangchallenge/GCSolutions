package main

import (
	"bytes"
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"golang.org/x/crypto/nacl/box"
)

type secReader struct {
	io.Reader
	key   *[32]byte // precomputed shared key
	nonce [24]byte
	box   [32 * 1024]byte // box with max message size
	plain *bytes.Reader   // contains plaintext after box is opened
}

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	var key [32]byte
	box.Precompute(&key, pub, priv)
	return &secReader{Reader: r, key: &key, plain: new(bytes.Reader)}
}

func (sr *secReader) Read(p []byte) (int, error) {
	if sr.plain.Len() > 0 {
		return sr.plain.Read(p)
	}

	_, err := sr.Reader.Read(sr.nonce[:])
	if err != nil {
		return 0, err
	}

	bn, err := sr.Reader.Read(sr.box[:])
	if err != nil {
		return 0, err
	}

	plain, ok := box.OpenAfterPrecomputation(nil, sr.box[:bn], &sr.nonce, sr.key)
	if !ok {
		return 0, fmt.Errorf("could not open box")
	}

	sr.plain = bytes.NewReader(plain)
	return sr.plain.Read(p)
}

type secWriter struct {
	io.Writer
	key   *[32]byte
	nonce [24]byte
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	var key [32]byte
	box.Precompute(&key, pub, priv)
	return &secWriter{Writer: w, key: &key}
}

func (sw *secWriter) Write(p []byte) (int, error) {
	// construct a random nonce
	n, err := rand.Read(sw.nonce[:])
	if err != nil || n != 24 {
		return 0, fmt.Errorf("could not create nonce:%s", err)
	}

	bx := box.SealAfterPrecomputation(nil, p, &sw.nonce, sw.key)
	b := append(sw.nonce[:], bx...)
	wn, err := sw.Writer.Write(b)
	if err != nil {
		return 0, err
	}

	// todo: what if not all of b is written: wn < b
	_ = wn
	return len(p), nil
}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	// receive server public key
	var servPub [32]byte
	if n, err := conn.Read(servPub[:]); err != nil || n != 32 {
		return conn, err
	}

	// send public key
	if n, err := conn.Write(pub[:]); err != nil || n != 32 {
		return conn, err
	}

	sr := NewSecureReader(conn, priv, &servPub)
	sw := NewSecureWriter(conn, priv, &servPub)

	rwc := &secReadWriteCloser{Reader: sr, Writer: sw, Closer: conn}
	return rwc, nil
}

type secReadWriteCloser struct {
	io.Reader
	io.Writer
	io.Closer
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	for {
		c, err := l.Accept()
		if err != nil {
			return fmt.Errorf("serve connection died:%s", err)
		}

		if err := serveConn(c); err != nil {
			return err
		}
	}

	return nil
}

func serveConn(c net.Conn) error {
	defer c.Close()

	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return fmt.Errorf("could not generate key: %s", err)
	}

	n, err := c.Write(pub[:])
	if err != nil || n != 32 {
		return fmt.Errorf("handshake failed, could not send public key:%s", err)
	}

	var clientPub [32]byte
	if n, err := c.Read(clientPub[:]); n != 32 || err != nil {
		return fmt.Errorf("handshake failed, did not receive public key:%s", err)
	}

	sr := NewSecureReader(c, priv, &clientPub)
	sw := NewSecureWriter(c, priv, &clientPub)

	if _, err := io.Copy(sw, sr); err != nil {
		return fmt.Errorf("serve copy failed: %s", err)
	}

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
