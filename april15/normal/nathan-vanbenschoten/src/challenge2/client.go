package main

import (
	"challenge2/secureio"
	"crypto/rand"
	"fmt"
	"io"
	"net"

	"golang.org/x/crypto/nacl/box"
)

// Dial generates a private/public key pair, connects to the server,
// performs a public-key handshake, and returns a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("error generating public and private key: %v", err)
	}

	c, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	if err := sendPublic(c, pub); err != nil {
		return nil, err
	}
	peersPub, err := receivePublic(c)
	if err != nil {
		return nil, err
	}

	secureConn := secureio.NewSecureConnection(c, priv, peersPub)
	return secureConn, nil
}
