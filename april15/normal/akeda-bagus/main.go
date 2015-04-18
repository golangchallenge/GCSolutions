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

const (
	nonceLen = 24    // Length of nonce bytes.
	maxData  = 32768 // Maximum bytes of message.
)

// header represents header in encrypted message. It contains one time nonce
// and length of encrypted message. This header is encoded in binary when
// transmitted through IO.
type header struct {
	Nonce      [nonceLen]byte
	MessageLen uint16
}

// NewSecureReader instantiates a new SecureReader.
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	return &SecureReader{
		r:    r,
		pub:  pub,
		priv: priv,
	}
}

// SecureReader represents secure reader in which input stream is encrypted.
// It implements io.Reader.
type SecureReader struct {
	r    io.Reader
	pub  *[32]byte // Peers public key.
	priv *[32]byte // Reader's private key.
}

// Read reads encrypted message from input stream r.
func (sr *SecureReader) Read(p []byte) (n int, err error) {
	var (
		h  header // Header (nonce + message length)
		d  []byte // Decrypted message
		ok bool
	)

	// Read header.
	if err = binary.Read(sr.r, binary.BigEndian, &h); err != nil {
		return
	}

	// Read encrypted message.
	encrypted := make([]byte, h.MessageLen)
	if _, err = sr.r.Read(encrypted); err != nil {
		return
	}

	// Decrypts encrypted message.
	if d, ok = box.Open(d, encrypted, &h.Nonce, sr.pub, sr.priv); !ok {
		return n, errors.New("Unable to decrypt encrypted message")
	}
	n = copy(p, d)

	return
}

// NewSecureWriter instantiates a new SecureWriter.
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	return &SecureWriter{w, pub, priv}
}

// SecureWriter represents secure writer in which message will be encrypted
// using writer's private key and peers public key using NaCl.
type SecureWriter struct {
	w    io.Writer
	pub  *[32]byte // Peers public key.
	priv *[32]byte // Sender private key.
}

// Write encrypts plaintext msg into w using peers public key and writer's
// private key.
func (sw *SecureWriter) Write(msg []byte) (n int, err error) {
	var (
		h         header
		encrypted []byte
	)

	// Generate nonce.
	if _, err = rand.Reader.Read(h.Nonce[:]); err != nil {
		return
	}

	// Encrypt the message.
	encrypted = box.Seal(encrypted, msg, &h.Nonce, sw.pub, sw.priv)

	// Length of encrypted message.
	h.MessageLen = uint16(len(encrypted))

	// Encode nonce and encrypted message length message with binary so that
	// reader know the length of the encrypted message.
	if err = binary.Write(sw.w, binary.BigEndian, h); err != nil {
		return
	}

	// Write the encrypted message.
	return sw.w.Write(encrypted)
}

// Dial generates a private/public key pair, connects to the server, perform
// the handshake and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	c := &client{
		conn: conn,
	}

	if err := c.exchangeKey(); err != nil {
		return nil, err
	}

	return c, nil
}

// client represents dialer that makes connection with secure echo server. It
// implements io.ReadWriteCloser.
type client struct {
	conn net.Conn
	r    io.Reader
	w    io.Writer
}

// exchangeKey exchanges client's public with server's public key. Client will
// sends public key first then server will replies its public key. Once public
// keys are set, secure reader and writer are instantiated so that client can
// sends encrypted message and decrypts the echo from the server.
func (c *client) exchangeKey() error {
	// Generate client's private and public keys.
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}

	// Sends client's public key to server.
	if _, err := c.conn.Write(pub[:]); err != nil {
		return err
	}

	// Gets server's public key from server.
	var serverPub [32]byte
	if _, err := c.conn.Read(serverPub[:]); err != nil {
		return err
	}

	// Sets secure reader and writer using server's public key.
	c.r = NewSecureReader(c.conn, priv, &serverPub)
	c.w = NewSecureWriter(c.conn, priv, &serverPub)

	return nil
}

// Read reads encrypted message from connection and decypts the message into b.
func (c *client) Read(b []byte) (n int, err error) {
	return c.r.Read(b)
}

// Write encrypts msg and sends encryped msg to connection.
func (c *client) Write(msg []byte) (n int, err error) {
	return c.w.Write(msg)
}

// Close closes the connection.
func (c *client) Close() error {
	return c.conn.Close()
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) (err error) {
	s := &server{
		l: l,
	}
	s.pub, s.priv, err = box.GenerateKey(rand.Reader)
	if err != nil {
		return
	}

	var conn net.Conn
	for {
		conn, err = s.l.Accept()
		if err != nil {
			break
		}
		go s.handleConn(conn)
	}

	return
}

// server represents secure echo server.
type server struct {
	l    net.Listener
	pub  *[32]byte
	priv *[32]byte
}

// handleConn handles incoming connection. handleConn runs on its own goroutine.
func (s *server) handleConn(c net.Conn) {
	defer c.Close()

	// Exchange public key with client.
	clientKey, err := s.exchangeKey(c)
	if err != nil {
		log.Fatal(err)
	}

	// Echo encrypted message with server's public key.
	if err := s.echo(c, clientKey); err != nil {
		log.Fatal(err)
	}
}

// exchangeKey exchanges public key with client. Server will reads client's
// public key and sends back server's public key.
func (s *server) exchangeKey(c net.Conn) (*[32]byte, error) {
	// Gets client public key.
	var clientPub [32]byte
	if _, err := c.Read(clientPub[:]); err != nil {
		return nil, err
	}

	// Sends back server's public key.
	if _, err := c.Write(s.pub[:]); err != nil {
		return nil, err
	}

	return &clientPub, nil
}

// echo decrypts message from client, encrypts the message with client's public
// key and server's private key, and sends back the encrypted message to client.
func (s *server) echo(c net.Conn, pub *[32]byte) error {
	var (
		d   [maxData]byte // Decrypted message from client
		n   int
		err error
	)

	// Decrypt message from client.
	n, err = NewSecureReader(c, s.priv, pub).Read(d[:])
	if err != nil {
		return err
	}

	// Encrypt message using client's public key.
	if _, err = NewSecureWriter(c, s.priv, pub).Write(d[:n]); err != nil {
		return err
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
