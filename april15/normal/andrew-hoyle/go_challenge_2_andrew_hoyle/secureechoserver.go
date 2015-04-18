package main

import (
	"crypto/rand"
	"io"
	"net"

	"golang.org/x/crypto/nacl/box"
)

func serverHandshake(c net.Conn, serPub *[32]byte) (*[32]byte, error) {
	// get client pub key
	b := make([]byte, 32)
	var cliPub [32]byte
	n, err := c.Read(b)
	if err != nil {
		return nil, err
	}
	copy(cliPub[:], b[:n])

	// send client pub key
	_, err = c.Write(serPub[:])
	if err != nil {
		return nil, err
	}

	return &cliPub, nil
}

func echoMsg(s io.ReadWriteCloser) error {
	b2 := make([]byte, 1024)
	n, err := s.Read(b2)
	if err != nil {
		return err
	}

	n, err = s.Write(b2[:n])
	if err != nil {
		return err
	}

	return nil
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	defer l.Close()

	serPub, serPriv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}

	// wait for connection
	conn, err := l.Accept()
	if err != nil {
		return err
	}

	cliPub, err := serverHandshake(conn, serPub)
	if err != nil {
		return err
	}

	s := NewSecureReadWriteCloser(conn, serPriv, cliPub)
	defer s.Close()

	err = echoMsg(s)
	if err != nil {
		return err
	}

	return nil
}
