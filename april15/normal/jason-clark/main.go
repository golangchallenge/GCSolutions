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

var (
	// ErrKeyExchange indicates an error exchanging the public keys
	ErrKeyExchange = errors.New("could not exchange public keys")
	// ErrKeyGeneration indicates an error generating public and private key pairs
	ErrKeyGeneration = errors.New("could not generate encryption keys")
	// ErrDecryption indicates an error decrypting a message
	ErrDecryption = errors.New("could not decrypt received message")
	// ErrNonceWrite indicates an error sending nonce value for message
	ErrNonceWrite = errors.New("could not send nonce value")
	// ErrNonceRead indicates an error reading nonce value for message
	ErrNonceRead = errors.New("could not read nonce value")
)

// SecureReader implements NaCl box encryption for an underlying io.Reader.
type SecureReader struct {
	r     io.Reader
	key   *[32]byte
	nonce *[24]byte
}

// Read reads from the underlying io.Reader and decrypts the contents into p.
func (r *SecureReader) Read(p []byte) (int, error) {
	// Each message starts with a nonce.  Only read the nonce once.
	if r.nonce == nil {
		var nonce [24]byte
		if _, err := io.ReadFull(r.r, nonce[:]); err != nil {
			return 0, ErrNonceRead
		}
		r.nonce = &nonce
	}

	// Read message from underlying Reader
	buf := make([]byte, len(p)+box.Overhead)
	n, err := r.r.Read(buf)
	if n <= 0 {
		return n, err
	}

	// Decrypt new message
	m, success := box.OpenAfterPrecomputation(nil, buf[:n], r.nonce, r.key)
	if !success {
		return 0, ErrDecryption
	}
	return copy(p, m), err
}

// SecureWriter implements NaCl box encryption for an underlying io.Writer.
type SecureWriter struct {
	w     io.Writer
	key   *[32]byte
	nonce *[24]byte
}

// Write encrypts the contents of p and writes it to the underlying io.Writer.
func (w *SecureWriter) Write(p []byte) (int, error) {
	// Each message starts with a generated nonce. Only write the nonce once.
	if w.nonce == nil {
		var nonce [24]byte
		if _, err := io.ReadFull(io.TeeReader(rand.Reader, w.w), nonce[:]); err != nil {
			return 0, ErrNonceWrite
		}
		w.nonce = &nonce
	}

	// encrypted and send message
	_, err := w.w.Write(box.SealAfterPrecomputation(nil, p, w.nonce, w.key))
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

// NewSecureConn instantiates a new io.ReadWriteCloser backed by a
// SecureReader and SecureWriter with public keys already exchanged.
func NewSecureConn(conn net.Conn) (io.ReadWriteCloser, error) {
	// Generate random key pair
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, ErrKeyGeneration
	}

	// Send public key
	if _, err := conn.Write(pub[:]); err != nil {
		return nil, ErrKeyExchange
	}

	// Read other side's public key
	var otherPub [32]byte
	if _, err := io.ReadFull(conn, otherPub[:]); err != nil {
		return nil, ErrKeyExchange
	}

	var key [32]byte
	box.Precompute(&key, &otherPub, priv)

	return struct {
		io.Reader
		io.Writer
		io.Closer
	}{
		Reader: &SecureReader{r: conn, key: &key},
		Writer: &SecureWriter{w: conn, key: &key},
		Closer: conn,
	}, nil
}

// NewSecureReader instantiates a new SecureReader.
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	var key [32]byte
	box.Precompute(&key, pub, priv)
	return &SecureReader{r: r, key: &key}
}

// NewSecureWriter instantiates a new SecureWriter.
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	var key [32]byte
	box.Precompute(&key, pub, priv)
	return &SecureWriter{w: w, key: &key}
}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return NewSecureConn(conn)
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	conn, err := l.Accept()
	if err != nil {
		return err
	}
	defer conn.Close()
	rw, err := NewSecureConn(conn)
	if err != nil {
		return err
	}
	n, err := io.Copy(io.MultiWriter(os.Stdout, rw), rw)
	if n > 0 {
		os.Stdout.Write([]byte("\n"))
	}
	return err
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
