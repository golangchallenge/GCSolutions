package main

import (
	"crypto/rand"
	"io"
	"log"
	"net"

	"golang.org/x/crypto/nacl/box"
)

// handleConnection handles a net.Conn for the life of the connection.
func handleConnection(conn net.Conn) {
	defer conn.Close()
	// Setup channels for incoming data and any errors.
	dataChan := make(chan []byte)
	errorChan := make(chan error)
	// Generate our public and private keys
	pubKey, privKey, err := box.GenerateKey(rand.Reader)
	if err != nil {
		log.Println("error generating keys:", err)
		return
	}
	// Read the clients/peers public key
	var peersPublicKey [32]byte
	_, err = conn.Read(peersPublicKey[:])
	if err != nil {
		log.Println("error reading pubkey:", err)
		return
	}
	// Send our public Key
	_, err = conn.Write(pubKey[:])
	if err != nil {
		log.Println("error writing pubkey:", err)
		return
	}
	// Set up Secure writers
	sr := NewSecureReader(conn, privKey, &peersPublicKey)
	sw := NewSecureWriter(conn, privKey, &peersPublicKey)

	// Start a goroutine to read from our net connection
	go func(dataChan chan []byte, errorChan chan error) {
		for {
			// Big buffer that can hold our max message size of 32KB.
			data := make([]byte, 32768)
			// Read the data...
			n, err := sr.Read(data)
			// And then shorten the buffer to only N bytes long, which is how many bytes we read.
			data = data[:n]
			if err != nil {
				// send an error to the error channel if it's encountered
				errorChan <- err
				continue
			}
			// And at last send the data into the read channel.
			dataChan <- data
		}
	}(dataChan, errorChan)

	// connection state machine that will last for the entire connection
	for {
		select {
		// This case means we recieved data on the connection
		case data := <-dataChan:
			// echo it back
			sw.Write(data[:])
		// This case means we got an error and the goroutine has finished
		case err := <-errorChan:
			// We want to print the error, but a disconnecting client will cause a EOF, so we can omit those.
			if err != io.EOF {
				log.Println(err)
			}
			// Exit the loop, this connection is done.
			break
		}
	}
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go handleConnection(conn)
	}
	// This will never be executed, we could prevent this by using select on a quit channel passed into Serve to detect when the program wants to quit.
	return nil
}
