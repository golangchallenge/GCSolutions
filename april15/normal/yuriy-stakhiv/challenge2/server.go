package main

import (
	"crypto/rand"
	"golang.org/x/crypto/nacl/box"
	"io"
	"log"
	"net"
)

// EchoServer is a simple echo server
type EchoServer struct {
	pub  *[32]byte
	priv *[32]byte
}

// NewEchoServer is an EchoServer constructor
func NewEchoServer() (*EchoServer, error) {
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	server := &EchoServer{pub, priv}
	return server, nil
}

// Serve listens and serves using provided net.Listener
func (s *EchoServer) Serve(l net.Listener) error {
	for {
		c, err := l.Accept()
		if err != nil {
			return err
		}
		go s.handleRequest(c)
	}
}

func (s *EchoServer) handleRequest(c net.Conn) {
	defer c.Close()
	buf := make([]byte, 32000)

	// Create secure connection
	sConn := NewSecureConn(c)
	// Init handshake
	err := sConn.ServerHandshake(s.pub, s.priv)
	if err != nil {
		log.Println("Failed handshake: ", err)
		return
	}

	for {
		rn, err := c.Read(buf)
		if err != nil {
			if err == io.EOF {
				return
			}
			log.Println("Error reading: ", err)
			return
		}
		_, err = c.Write(buf[:rn])
		if err != nil {
			if err == io.EOF {
				log.Println("Connection closed unexpectedly")
				return
			}
			log.Println("Error writing: ", err)
			return
		}

	}
}
