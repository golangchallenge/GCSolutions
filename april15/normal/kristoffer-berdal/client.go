package main

import (
	"crypto/rand"
	"io"
	"net"

	"golang.org/x/crypto/nacl/box"
)

// Dial generates a private/public key pair,
// connects to the server, performs the handshake
// and return a reader/writer.
// Flow: Dial server->send our pub key->receive servers pub key
func Dial(addr string) (io.ReadWriteCloser, error) {
	// Generate our keys
	pubKey, privKey, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	var peersPublicKey [32]byte

	// Dial the server
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	// Write our public key to client
	_, err = conn.Write(pubKey[:])
	if err != nil {
		return nil, err
	}
	// Read clients/peers public key
	_, err = conn.Read(peersPublicKey[:])
	if err != nil {
		return nil, err
	}
	// If we got no error then all of the public key was read
	// Set up secure reader and writer
	sr := NewSecureReader(conn, privKey, &peersPublicKey)
	sw := NewSecureWriter(conn, privKey, &peersPublicKey)
	return SecureReadWriteCloser{sr, sw, conn}, nil
}
