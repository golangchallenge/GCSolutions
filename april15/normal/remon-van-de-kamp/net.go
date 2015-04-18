package main

import (
	"crypto/rand"
	"encoding/binary"
	"golang.org/x/crypto/nacl/box"
	"io"
	"log"
	"net"
)

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	netConn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	var theirPub [32]byte
	if err = binary.Read(netConn, binary.LittleEndian, &theirPub); err != nil {
		return nil, err
	}

	ourPub, ourPriv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	if _, err = netConn.Write(ourPub[:]); err != nil {
		return nil, err
	}

	return NewSecureConnection(netConn, ourPriv, &theirPub), nil
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	for {
		netConn, err := l.Accept()
		if err != nil {
			return err
		}

		// Handle connection in a goroutine so we are able
		// to handle multiple connections simultaniously.
		go func(netConn net.Conn) {
			defer netConn.Close()

			// create a public/private keypair for each
			// connection so that when MITM occurs they
			// can only listen in on one conversation, and
			// not all of them
			ourPub, ourPriv, err := box.GenerateKey(rand.Reader)
			if err != nil {
				log.Printf("Server failed to generate key pair: %s", err)
				return
			}

			err = binary.Write(netConn, binary.LittleEndian, ourPub)
			if err != nil {
				log.Printf("Server failed to send public key to client: %s", err)
				return
			}

			var theirPub [32]byte
			err = binary.Read(netConn, binary.LittleEndian, &theirPub)
			if err != nil {
				log.Printf("Error receiving public key from client: %s", err)
				return
			}

			conn := NewSecureConnection(netConn, ourPriv, &theirPub)

			data := make([]byte, 1<<15)
			n, err := conn.Read(data)

			if err != nil {
				log.Printf("Error receiving data from client: %s (read %d bytes)", err, n)
				return
			}
			conn.Write(data[:n])
			conn.Close()
		}(netConn)
	}
	return nil
}
