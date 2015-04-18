package main

import (
	"crypto/rand"
	"fmt"
	"golang.org/x/crypto/nacl/box"
	"io"
	"log"
	"net"
)

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {

	fmt.Println("Starting server...")
	// Generate a public/private keypair for server
	spub, spriv, _ := box.GenerateKey(rand.Reader)

	// Accept the connection then handle it
	for {
		fmt.Println("Waiting for connection...")
		conn, err := l.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		// Go to a go routine, so we can accept more connections
		handleSecureConnection(conn, spub, spriv)
	}
}

func handleSecureConnection(conn net.Conn, spub, spriv *[32]byte) {

	// Send server's public key to client
	key := [32]byte(*spub)
	_, err := conn.Write(key[:])
	if err != nil {
		log.Println("Could not send public key to client...")
		log.Fatal(err)
	}

	// Read client's public key
	var cpub [32]byte
	_, err = io.ReadAtLeast(conn, cpub[:], 32)
	if err != nil {
		log.Println("Could not get client's public key...")
		log.Fatal(err)
	}

	// Create reader/writer to talk to client
	r := NewSecureReader(conn, spriv, &cpub)
	w := NewSecureWriter(conn, spriv, &cpub)

	// Read from the client
	buf := make([]byte, 4096)
	n, err := r.Read(buf)
	if err != nil || n == 0 {
		conn.Close()
	}
	// Echo back to the client
	n, err = w.Write(buf[:n])
	if err != nil {
		conn.Close()
	}
	log.Printf("%s disconnected.\n", conn.RemoteAddr().String())
}
