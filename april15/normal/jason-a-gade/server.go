package main

import (
	"crypto/rand"
	"fmt"
	"io"
	"net"

	"golang.org/x/crypto/nacl/box"
)

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	peer, err := exchangeKeys(conn, pub)
	if err != nil {
		return nil, err
	}
	secure := NewSecureConn(conn, priv, peer)
	return secure, nil
}

// Serve starts a secure echo server on the given listener. Sending an empty
// message will quit the server.
func Serve(l net.Listener) error {

	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		defer conn.Close()

		pub, priv, err := box.GenerateKey(rand.Reader)
		if err != nil {
			return err
		}

		peer, err := exchangeKeys(conn, pub)
		if err != nil {
			return err
		}
		secure := NewSecureConn(conn, priv, peer)

		// If nothing is copied then the server will end. This allows the client to
		// send an empty message to kill the server.
		if n, err := io.Copy(secure, secure); n == 0 {
			if err != nil {
				return err
			}
			return fmt.Errorf("Server ending.")
		}
	}
}

// exchangeKeys exchanges keys between the client and the server. The same function
// can be used for both because with such small pieces of data, the order of Write
// and Read should not matter.
func exchangeKeys(conn net.Conn, pub *[keySize]byte) (*[keySize]byte, error) {

	peer := new([keySize]byte)

	_, err := conn.Write(pub[:])
	if err != nil {
		return nil, err
	}

	_, err = conn.Read(peer[:])
	if err != nil {
		return nil, err
	}

	return peer, nil
}
