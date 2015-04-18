package main

import (
	"crypto/rand"
	"io"
	"net"

	"golang.org/x/crypto/nacl/box"
)

func clientHandshake(c net.Conn, cliPub *[32]byte) (*[32]byte, error) {
	// send pub key to server
	n, err := c.Write(cliPub[:])
	if err != nil {
		return nil, err
	}

	// get pub key from server
	b := make([]byte, 32)
	var serPub [32]byte
	n, err = c.Read(b)
	if err != nil {
		return nil, err
	}
	copy(serPub[:], b[:n])

	return &serPub, nil
}

// Dial generates a private/public key pair,
// connects to the server, performs the handshake
// and returns a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	cliPub, cliPriv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	serPub, err := clientHandshake(conn, cliPub)
	if err != nil {
		return nil, err
	}

	s := NewSecureReadWriteCloser(conn, cliPriv, serPub)

	return s, nil
}
