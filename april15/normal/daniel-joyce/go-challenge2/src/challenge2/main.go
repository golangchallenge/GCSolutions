package main

import (
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"golang.org/x/crypto/nacl/box"
)

const (
	keySize   int = 32
	nonceSize int = 24
	msgSize   int = (1 << 15) + nonceSize
)

// SecureReader implements io.Reader by reading and decrypting
// messages using a public/private key pair.
type SecureReader struct {
	r    io.Reader
	priv *[keySize]byte
	pub  *[keySize]byte
}

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, pub *[keySize]byte) io.Reader {
	return SecureReader{r, priv, pub}
}

func (sr SecureReader) Read(p []byte) (int, error) {
	enc := make([]byte, msgSize)
	n, err := sr.r.Read(enc)
	if err != nil {
		return n, err
	}

	if n < nonceSize {
		return n, errors.New("Message is too small. Is there a nonce?")
	}

	var nonce [nonceSize]byte
	copy(nonce[:], enc[:nonceSize])

	tmp, ok := box.Open(p[:0], enc[nonceSize:n], &nonce, sr.pub, sr.priv)
	if !ok {
		return n, errors.New("Unable to decrypt message.")
	}

	return len(tmp), nil
}

// SecureWriter implements io.Writer by writing and encrypting
// messages using a public/private key pair.
type SecureWriter struct {
	w    io.Writer
	priv *[keySize]byte
	pub  *[keySize]byte
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[keySize]byte) io.Writer {
	return SecureWriter{w, priv, pub}
}

func (sw SecureWriter) Write(p []byte) (int, error) {
	var nonce [nonceSize]byte
	_, err := rand.Read(nonce[:])
	if err != nil {
		return 0, err
	}

	enc := append(nonce[:], box.Seal(nil, p, &nonce, sw.pub, sw.priv)...)

	return sw.w.Write(enc)
}

// SecureConn reads and writes encrypted messages using a public/private key pair.
type SecureConn struct {
	c io.Closer
	r io.Reader
	w io.Writer
}

// NewSecureConn instantiates a secure connection by wrapping
// an existing io.ReadWriteCloser.
func NewSecureConn(c io.ReadWriteCloser, localPub, remotePub, localPriv *[keySize]byte) io.ReadWriteCloser {
	r := NewSecureReader(c, localPriv, remotePub)
	w := NewSecureWriter(c, localPriv, remotePub)
	return SecureConn{c, r, w}
}

func (s SecureConn) Read(b []byte) (int, error) {
	return s.r.Read(b)
}

func (s SecureConn) Write(b []byte) (int, error) {
	return s.w.Write(b)
}

// Close closes the underlying io.ReadWriteCloser.
func (s SecureConn) Close() error {
	return s.c.Close()
}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	localPub, localPriv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	// Send my public key.
	_, err = conn.Write(localPub[:])
	if err != nil {
		return nil, err
	}

	// Receive the servers public key.
	var remotePub [keySize]byte
	_, err = conn.Read(remotePub[:])
	if err != nil {
		return nil, err
	}

	return NewSecureConn(conn, localPub, &remotePub, localPriv), nil
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	// Generate public and private keys.
	localPub, localPriv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		log.Println(err)
		return
	}

	// Receive the clients public key.
	var remotePub [keySize]byte
	_, err = conn.Read(remotePub[:])
	if err != nil {
		log.Println(err)
		return
	}

	// Send my public key.
	_, err = conn.Write(localPub[:])
	if err != nil {
		log.Println(err)
		return
	}

	secConn := NewSecureConn(conn, localPub, &remotePub, localPriv)

	buf := make([]byte, msgSize)
	n, err := secConn.Read(buf)
	if err != nil {
		log.Println(err)
		return
	}

	_, err = secConn.Write(buf[:n])
	if err != nil {
		log.Println(err)
		return
	}
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
