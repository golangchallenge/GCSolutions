package main

import (
	"crypto/rand"
	"golang.org/x/crypto/nacl/box"
	"io"
	"log"
	"net"
)

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {

	// Generate key pair for client
	cpub, cpriv, _ := box.GenerateKey(rand.Reader)

	// Attempt to connect to server
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Println("Could not make connection with server...")
		log.Fatal(err)
	}

	// Get server's public key
	var spub [32]byte
	_, err = io.ReadAtLeast(conn, spub[:], 32)
	if err != nil {
		log.Println("Could not get server's public key...")
		log.Fatal(err)
	}

	// Send client's public key
	key := [32]byte(*cpub)
	_, err = conn.Write(key[:])
	if err != nil {
		log.Println("Could not send public key to server...")
		log.Fatal(err)
	}

	// Return client a reader/writer
	r := NewSecureReader(conn, cpriv, &spub)
	w := NewSecureWriter(conn, cpriv, &spub)
	c := NewSecureCloser(conn)
	srwc := NewSecureReadWriteCloser(r, w, c)
	return srwc, err
}
