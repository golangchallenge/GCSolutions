/*
Package sio provides a simple wrapper for golang.org/x/crypto/nacl. It enables
easy setup of encrypted communication by encapsulating key generation and
exchange, and implementing io.Reader and io.Writer interfaces.

This package was written while learning Go. It is not meant for real world use.
*/
package sio

import (
	"crypto/rand"
	"io"
	"net"

	"golang.org/x/crypto/nacl/box"
)

// A Conduit provides basic encapsulation of the golang.org/x/crypto/nacl
// library. A Conduit should be created using the NewConduit function.
//
// Conduit's interfaces all interact with the underlying net.Conn provided to
// NewConduit. Read() will retrieve and decrypt incoming data through a
// SecureReader. Write() will encrypt and write outgoing data by way of a
// SecureWriter. Calling a Conduit's Close method will close the associated
// network connection.
type Conduit struct {
	io.Reader
	io.Writer
	io.Closer
	conn      net.Conn
	priv, pub *[32]byte
	peerpub   *[32]byte
}

// initConduit generates a key pair for the conduit's SecureReader and
// SecureWriter, sends the public key and waits to receive its peer's
// public key
func (c *Conduit) initConduit() (err error) {
	// TODO: generate errors for len != 32 when sending and receiving keys
	var len int

	if c.pub, c.priv, err = box.GenerateKey(rand.Reader); err != nil {
		return
	}

	// send our public key to peer
	if len, err = c.conn.Write(c.pub[:]); err != nil || len != 32 {
		return
	}

	// receive and store peer's public key
	c.peerpub = new([32]byte)

	if len, err = c.conn.Read(c.peerpub[:]); err != nil || len != 32 {
		return
	}

	return
}

// NewConduit will initialize and return a Conduit tied to the provided
// net.Conn. There should be a Conduit on the other end of the network
// connection to complete the required public key exchange.
func NewConduit(conn net.Conn) (c *Conduit, err error) {
	c = new(Conduit)

	c.conn = conn
	c.initConduit()

	c.Reader = NewSecureReader(conn, c.priv, c.peerpub)
	c.Writer = NewSecureWriter(conn, c.priv, c.peerpub)
	c.Closer = conn

	return
}
