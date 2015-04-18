package main

import (
	"crypto/rand"
	"io"
	"net"

	"golang.org/x/crypto/nacl/box"
)

// SecureReadWriter create a secure network connection.
type SecureReadWriter struct {
	conn net.Conn
	io.Reader
	io.Writer
}

// Close implements io.Closer to make SecureReadWriter
// as a io.ReadWriteCloser
func (s *SecureReadWriter) Close() error {
	return s.conn.Close()
}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	public, private, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	var serverKey [32]byte
	_, err = io.ReadFull(conn, serverKey[:])
	if err != nil {
		return nil, err
	}
	_, err = conn.Write(public[:])
	if err != nil {
		return nil, err
	}

	return &SecureReadWriter{
		conn,
		NewSecureReader(conn, private, &serverKey),
		NewSecureWriter(conn, private, &serverKey),
	}, nil
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	public, private, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go handleConnection(conn, public, private)
	}
	return nil
}

// handleConnection serve a single client connection.
func handleConnection(conn net.Conn, public, private *[32]byte) {
	defer conn.Close()

	_, err := conn.Write(public[:])
	if err != nil {
		return
	}

	var peersPublicKey [32]byte
	_, err = io.ReadFull(conn, peersPublicKey[:])
	if err != nil {
		return
	}

	rw := &SecureReadWriter{
		conn,
		NewSecureReader(conn, private, &peersPublicKey),
		NewSecureWriter(conn, private, &peersPublicKey),
	}

	message := make([]byte, 32*1024)
	for {
		n, err := rw.Read(message)
		if err != nil {
			return
		}
		_, err = rw.Write(message[:n])
		if err != nil {
			return
		}
	}
}
