package secure

import (
	"crypto/rand"
	"encoding/binary"
	"net"

	"golang.org/x/crypto/nacl/box"
)

// A Conn wraps a new.Conn and encrypts the exchange using NaCl.
type Conn struct {
	w       *Writer
	r       *Reader
	closeFn func() error
}

// NewConn generates a keypair, performs handshake to exchange public keys with server and
// returns an encrypted connection wrapper.
func NewConn(conn net.Conn) (client *Conn, err error) {
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	if err := binary.Write(conn, binary.LittleEndian, *pub); err != nil {
		return nil, err
	}

	if err := binary.Read(conn, binary.LittleEndian, pub[:]); err != nil {
		return nil, err
	}

	return &Conn{
		w:       NewWriter(conn, priv, pub),
		r:       NewReader(conn, priv, pub),
		closeFn: conn.Close,
	}, nil
}

func (conn *Conn) Read(p []byte) (n int, err error) {
	return conn.r.Read(p)
}

func (conn *Conn) Write(p []byte) (n int, err error) {
	return conn.w.Write(p)
}

// Close closes underlying connection
func (conn *Conn) Close() error {
	return conn.closeFn()
}
