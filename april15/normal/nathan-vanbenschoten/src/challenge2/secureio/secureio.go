// Package secureio provides communication primitives that leverage NaCl to
// establish secure communication.
package secureio

import (
	"io"
	"net"
)

// NewSecureConnection instantiates a new io.ReadWriteCloser client that wraps
// a network connection and uses a secure reader and writer to ensure secure
// communication.
func NewSecureConnection(c net.Conn, priv, pub *[32]byte) io.ReadWriteCloser {
	return &secureConnection{
		NewSecureReader(c, priv, pub).(*secureReader),
		NewSecureWriter(c, priv, pub).(*secureWriter),
		c,
	}
}

// secureConnection wraps a secureReader and secureWriter, along with the
// network connection both io interfaces are attached to.
type secureConnection struct {
	*secureReader
	*secureWriter
	conn net.Conn
}

// Close closes the secureConnection's contained connection.
func (sc *secureConnection) Close() error {
	return sc.conn.Close()
}
