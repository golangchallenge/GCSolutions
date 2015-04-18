package main

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net"

	"golang.org/x/crypto/nacl/box"
)

// Dial generates a private/public key pair,
// connects to the server, perform the exchange
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	v(2).Printf("client: connected to %v", conn.RemoteAddr())

	cc, err := newClientConn(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return cc, nil
}

// newClientConn wraps the provided connections and enables secured
// duplex communication between the client and the server.
func newClientConn(conn net.Conn) (*clientConn, error) {
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	privK, pubK := (*key)(priv), (*key)(pub)

	v(2).Printf("client: generated private key %v, public key %v", privK, pubK)

	peersPub, err := exchange(conn, pubK)
	if err != nil {
		return nil, fmt.Errorf("key exchange failed with server %v: %v", conn.RemoteAddr(), err)
	}

	v(2).Printf("client: peer's public key is %v", peersPub)

	cc := &clientConn{
		NewSecureReader(conn, privK, peersPub),
		NewSecureWriter(conn, privK, peersPub),
		conn,
	}

	return cc, nil
}

// clientConn represents a connection to server.
type clientConn struct {
	io.Reader           // embedded secure reader
	secureW   io.Writer // secure writer
	io.Closer           // embedded conn

}

func (cc *clientConn) Write(p []byte) (int, error) {
	if len(p) > 32*1024 {
		return 0, errors.New("max message size supported is 32kb")
	}

	return cc.secureW.Write(p)
}
