package main

/*
golang-challenge/2
Jordan Wright <github.com/jordan-wright>

Design:

This challenge implements an echo server that communicates securely using
encryption provided by the Nacl library.

Communication is established via the Dial function, which returns a SecureConn
This SecureConn is an io.ReadWriteCloser that performs io using a SecureReader
and SecureWriter for decryption/encryption, respectively.
*/

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

// ErrKeyExchangeFailed is thrown if the server
// fails to accept the given shared key.
var ErrKeyExchangeFailed = errors.New("failed to send key to server")

// ErrGenNonceFailed is thrown when we do not read enough
// bytes from the random reader
var ErrGenNonceFailed = errors.New("failed to read enough random bytes into the nonce")

// ErrInvalidKey is thrown if we do not receive 32 bytes on connection
var ErrInvalidKey = errors.New("recieved invalid key from client")

// ErrBoxOpenFailed is thrown if we cannot open the encrypted box
var ErrBoxOpenFailed = errors.New("failed to open the encrypted box")

// KeyPair represents a public and private key pair.
type KeyPair struct {
	Public  [32]byte
	Private [32]byte
}

// generateNonce creates a random nonces.
// In some implementations, it might be worth making a
// counter, but for now we'll just use 24 random bytes
func generateNonce(p []byte) (int, error) {
	return rand.Read(p)
}

// SecureReader is a reader that reads from an io.Reader,
// and provides and encrypted Nacl stream to read from
type SecureReader struct {
	Reader    io.Reader
	KeyPair   KeyPair
	SharedKey [32]byte
	nonce     [24]byte
}

// NewSecureReader instantiates a new SecureReader.
// The shared key is precomputed to offer efficiency
// in key re-use across multiple messages.
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	var sK [32]byte
	var nonce [24]byte
	n, err := generateNonce(nonce[:])
	if err != nil {
		panic(err)
	}
	if n != len(nonce) {
		panic(ErrGenNonceFailed)
	}
	box.Precompute(&sK, pub, priv)
	kp := KeyPair{
		Public:  *pub,
		Private: *priv,
	}
	return &SecureReader{
		Reader:    r,
		KeyPair:   kp,
		SharedKey: sK,
		nonce:     nonce,
	}
}

// Read implements the Reader interface
// Need to read from the given slice and
// decrypt it using Nacl
func (s *SecureReader) Read(p []byte) (n int, err error) {
	// All messages shorter than 32kb
	data := make([]byte, 32*1024)
	// Read in the full data
	n, err = s.Reader.Read(data)
	if err != nil {
		return 0, err
	}
	data = data[:n]
	// Extract out the nonce
	for i := range s.nonce {
		s.nonce[i] = data[i]
	}
	// Extract out the encrypted message
	encrypted := data[len(s.nonce):n]
	// Decrypt the message
	decrypted, ok := box.OpenAfterPrecomputation(nil, encrypted, &s.nonce, &s.SharedKey)
	if !ok {
		return 0, ErrBoxOpenFailed
	}
	n = copy(p, decrypted)
	return n, nil
}

// SecureWriter implements the io.Writer interface.
// It writes a Nacl stream to the given writer.
// Both the Writer and the Reader will generate
// a shared key from the given private and public keys.
type SecureWriter struct {
	Writer    io.Writer
	KeyPair   KeyPair
	SharedKey [32]byte
	nonce     [24]byte
}

// NewSecureWriter instantiates a new SecureWriter.
// The shared key is precomputed to offer efficiency
// in key re-use across multiple messages.
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	var sK [32]byte
	var n [24]byte
	box.Precompute(&sK, pub, priv)
	// Initialize the first 16 bytes to random data
	_, err := rand.Read(n[:16])
	if err != nil {
		log.Println(err)
	}
	kp := KeyPair{
		Public:  *pub,
		Private: *priv,
	}
	return &SecureWriter{
		Writer:    w,
		KeyPair:   kp,
		SharedKey: sK,
		nonce:     n,
	}
}

// Write encrypts the given data and writes it to the given
// Writer interface
func (w *SecureWriter) Write(p []byte) (n int, err error) {
	generateNonce(w.nonce[:])
	encrypted := make([]byte, 1024)
	encrypted = box.SealAfterPrecomputation(nil, p, &w.nonce, &w.SharedKey)
	n, err = w.Writer.Write(append(w.nonce[:], encrypted...))
	return n, err
}

// SecureConn is a struct implementing
// the ReadWriteCloser interface
type SecureConn struct {
	Reader *SecureReader
	Writer *SecureWriter
	conn   io.ReadWriteCloser
}

// NewSecureConn returns an initialized secure connection object
// that satisfies the io.ReadWriteCloser interface
func NewSecureConn(pub, priv *[32]byte, c net.Conn) io.ReadWriteCloser {
	r := NewSecureReader(c, priv, pub)
	w := NewSecureWriter(c, priv, pub)
	return &SecureConn{
		Reader: r.(*SecureReader),
		Writer: w.(*SecureWriter),
		conn:   c,
	}
}

// Read receives data from the established TCP connection
// and decrypts it
func (s *SecureConn) Read(p []byte) (int, error) {
	return s.Reader.Read(p)
}

// Write encrypts the given data and sends it over the established
// TCP connection
func (s *SecureConn) Write(p []byte) (int, error) {
	return s.Writer.Write(p)
}

// Close closes the underlying TCP connection
func (s *SecureConn) Close() error {
	return s.conn.Close()
}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	// Generate the key pair
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	// Connect to the server
	c, err := net.Dial("tcp", addr)
	if err != nil {
		return c, err
	}
	// Send our public key
	n, err := c.Write(pub[:])
	if err != nil {
		log.Println(err)
	}
	// Get the public key from the server
	buff := make([]byte, 1024)
	n, err = c.Read(buff)
	if err != nil {
		log.Println(err)
		return c, ErrKeyExchangeFailed
	}
	var spub [32]byte
	if n != len(spub) {
		return c, ErrInvalidKey
	}
	buff = buff[:n]
	copy(spub[:], buff)
	sc := NewSecureConn(&spub, priv, c)
	// Return the reader/writer
	return sc, nil
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	defer l.Close()
	// Generate the server's public/private keys
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}
	for {
		c, err := l.Accept()
		if err != nil {
			return err
		}
		go echo(c, pub, priv)
	}
}

// Echoing function (used to handle each client connecion)
func echo(c net.Conn, pub, priv *[32]byte) {
	buff := make([]byte, 32*1024)
	var cpub [32]byte
	// Read the client's public key
	n, err := c.Read(buff)
	buff = buff[:n]
	if err != nil || n != len(cpub) {
		c.Close()
		log.Println(ErrInvalidKey)
		return
	}
	// Send the server public key to the client
	n, err = c.Write(pub[:])
	if err != nil {
		c.Close()
		return
	}
	copy(cpub[:], buff)
	conn := NewSecureConn(&cpub, priv, c)
	for {
		// Regrow our buffer to handle a max of 32kb
		buff = buff[:cap(buff)]
		// Read the data
		n, err = conn.Read(buff)
		if err != nil {
			c.Close()
			return
		}
		buff = buff[:n]
		log.Printf("%s\n", buff)
		// Echo back out the data
		n, err = conn.Write(buff)
		if err != nil {
			log.Print(err)
			return
		}
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
	if len(os.Args) < 3 {
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
	log.Printf("%s\n", buf[:n])
	err = conn.Close()
	if err != nil {
		log.Fatal(err)
	}
}
