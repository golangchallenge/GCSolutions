package main

import (
	"crypto/rand"
	"io"
	"net"
	"time"

	"golang.org/x/crypto/nacl/box"
)

type readWriteCloserImpl struct {
	io.Reader
	io.Writer
	io.Closer
}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (rwc io.ReadWriteCloser, err error) {
	// Generate private/public key pair;
	pub, priv, err := box.GenerateKey(rand.Reader)

	var c net.Conn
	if err == nil {
		// Connect to the server;
		c, err = net.Dial("tcp", addr)
	}

	if err == nil {
		// ClientHello: send public key;
		_, err = c.Write(pub[:])
	}

	// Give at most 15 seconds to the server for replying
	// with ServerHello.
	c.SetReadDeadline(time.Now().Add(15 * time.Second))

	// ServerHello: read server key;
	var otherPub [32]byte
	if err == nil {
		_, err = io.ReadFull(c, otherPub[:])
	}

	c.SetReadDeadline(time.Time{})

	// Setup the connection object.
	if err == nil {
		r := NewSecureReader(c, priv, &otherPub)
		w := NewSecureWriter(c, priv, &otherPub)
		rwc = &readWriteCloserImpl{Reader: r, Writer: w, Closer: c}
	}

	return
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) (err error) {
	for err == nil {
		var c net.Conn
		c, err = l.Accept()
		if err == nil {
			go serveEcho(c)
		}
	}
	return
}

func serveEcho(c net.Conn) {
	defer c.Close()

	var err error

	// 15 seconds at most to finish receiving ClientHello.
	c.SetReadDeadline(time.Now().Add(15 * time.Second))

	// ClientHello: Read public key;
	var otherPub [32]byte
	_, err = io.ReadFull(c, otherPub[:])

	// ClientHello is either received, or we are in error.
	c.SetReadDeadline(time.Time{})

	// Generate private/public key pair
	var pub, priv *[32]byte
	if err == nil {
		pub, priv, err = box.GenerateKey(rand.Reader)
	}

	// ServerHello: Send server public key
	if err == nil {
		_, err = c.Write(pub[:])
	}

	var r io.Reader
	var w io.Writer
	if err == nil {
		r = NewSecureReader(c, priv, &otherPub)
		w = NewSecureWriter(c, priv, &otherPub)
	}

	var buf [256]byte
	var n int
	for err == nil {
		// when reading data, receive encrypted data from client;
		c.SetReadDeadline(time.Now().Add(60 * time.Second))
		n, err = io.ReadAtLeast(r, buf[:], 1)

		// send encrypted data from server
		if err == nil {
			_, err = w.Write(buf[:n])
		}
	}

	// Dump the error message to the stream.
	// Reason: TestServe will not accept a connection closure
	// without any data.
	if err != io.EOF {
		c.Write([]byte(err.Error()))
	}
}
