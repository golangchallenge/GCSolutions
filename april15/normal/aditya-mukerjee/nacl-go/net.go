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
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	raddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTCP("tcp", nil, raddr)

	// Read the public key from the server
	spub := make([]byte, 32)
	n, err := conn.Read(spub)
	if err == nil && n != 32 {
		err = fmt.Errorf("expected 32-byte public key from server and received %d", n)
	}
	if err != nil {
		return nil, err
	}

	// Send the client public key to the server
	_, err = conn.Write(pub[:])
	if err != nil {
		return nil, err
	}

	var spuba [32]byte
	for i := 0; i < len(spub); i++ {
		spuba[i] = spub[i]
	}
	sr := NewSecureReader(conn, priv, &spuba)
	sw := NewSecureWriter(conn, priv, &spuba)

	sc := secureConnection{sr, sw, conn}

	return sc, err
}

type secureConnection struct {
	sr   io.Reader
	sw   io.Writer
	conn *net.TCPConn
}

func (sc secureConnection) Read(b []byte) (n int, err error) {
	n, err = sc.sr.Read(b)
	if err == io.EOF {
		err = nil
	}
	if err == nil && n == 0 {
		err = fmt.Errorf("read zero bytes")
	}
	return
}

func (sc secureConnection) Write(b []byte) (n int, err error) {
	n, err = sc.sw.Write(b)
	if err == io.EOF {
		err = nil
	}
	return
}
func (sc secureConnection) Close() error {
	return sc.conn.Close()
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	spub, spriv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}

	conn, err := l.Accept()
	if err != nil {
		return err
	}

	go func(c net.Conn) {
		defer c.Close()

		_, err := c.Write(spub[:])
		if err != nil {
			// TODO don't panic
			panic(err)
		}

		// puba holds the client's public key
		var puba [32]byte
		pub := make([]byte, 32)
		n, err := c.Read(pub)
		if err != nil {
			panic(err)
		}

		bts := make([]byte, MaxMessageBytes)
		for i := 0; i < len(pub); i++ {
			puba[i] = pub[i]
		}

		sr := NewSecureReader(c, spriv, &puba)
		sw := NewSecureWriter(c, spriv, &puba)

		n, err = sr.Read(bts)
		if err == nil && n == 0 {
			err = fmt.Errorf("server zero bytes from secure reader")
		}

		bts = bts[:n]

		if err != nil {
			panic(err)
		}

		n, err = sw.Write(bts)
		if err != nil {
			panic(err)
		}

	}(conn)

	return nil
}
