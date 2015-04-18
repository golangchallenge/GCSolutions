// Package main was created for the golang-challenge number 2
// (http://golang-challenge.com/go-challenge2/) by:
// Adrian Fletcher
// github.com/AdrianFletcher
// adrian@fletchtechnology.com.au

// Package main implements the tests found in main_test.go
// This package implements a handshake protocol for the boxcrypto
// package and uses the messenger package for client/server
// communication.
package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	crypto "github.com/AdrianFletcher/go-challenge/challenge_2/boxcrypto"
	"github.com/AdrianFletcher/go-challenge/challenge_2/messenger"
	"io"
	"log"
	"net"
	"os"
)

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	return crypto.NewSecureReader(r, priv, pub)
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	return crypto.NewSecureWriter(w, priv, pub)
}

// Handshake is the high level representation of the handshake message
// (which is a simple 32 byte slice) - expected to be a public key for
// encryption.
type Handshake struct {
	LocalPublicKey  [32]byte
	RemotePublicKey [32]byte
}

// SendHandshake sends the 32 byte public key only (in clear
// text). In a more robust system, the key should be exchanged more
// securely.
func (hs *Handshake) SendHandshake(w io.Writer) error {
	return binary.Write(w, binary.LittleEndian, hs.LocalPublicKey)
}

// GetHandshake receives the 32 byte public key only (in clear
// text). In a more robust system, the key should be exchanged more
// securely.
func (hs *Handshake) GetHandshake(r io.Reader) error {
	return binary.Read(r, binary.LittleEndian, &hs.RemotePublicKey)
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	// Generate new private/public key pair using a crypto secure random key
	pub, priv, err := crypto.GenerateKeyPair()
	if err != nil {
		return err
	}

	// Start a messenger server with our custom handshaker
	hs := &Handshake{
		LocalPublicKey: *pub,
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			return nil
		}

		// Serve the request with messenger
		go messenger.Serve(
			conn,
			hs,
			NewSecureWriter(conn, priv, &hs.RemotePublicKey),
			NewSecureReader(conn, priv, &hs.RemotePublicKey),
		)
	}
}

// Helper functions to satisfy the expected ReadWriteCloser interface
type cryptoReadWriteCloser struct {
	reader io.Reader
	writer io.Writer
}

func (c *cryptoReadWriteCloser) Read(p []byte) (n int, err error) {
	return c.reader.Read(p)
}
func (c *cryptoReadWriteCloser) Write(p []byte) (n int, err error) {
	return c.writer.Write(p)
}
func (c *cryptoReadWriteCloser) Close() error {
	return nil
}

// Dial generates a private/public key pair, connects to the server (using
// the messenger package) and returns a ReadWriteCloser per test specs.
func Dial(addr string) (io.ReadWriteCloser, error) {
	// Generate new private/public key pair using a crypto secure random key
	pub, priv, err := crypto.GenerateKeyPair()
	if err != nil {
		return nil, err
	}

	hs := &Handshake{
		LocalPublicKey: *pub,
	}

	// Dial a messenger with our custom handshaker. if it fails, send a blank message over
	// the connection (to satisfy test case TestSecureDial
	conn, err := messenger.Dial(addr, hs)
	if err != nil {
		return nil, err
	}

	// Create new type satisfying the ReadWriteCloser interface but with the crypto Reader/Writer
	return &cryptoReadWriteCloser{
		reader: NewSecureReader(conn, priv, &hs.RemotePublicKey),
		writer: NewSecureWriter(conn, priv, &hs.RemotePublicKey),
	}, nil
}

// The program can be started in either server mode or client mode.
// Server mode:
// main.exe -l <port number>
//
// Client mode:
// main.exe <port> <message>
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
