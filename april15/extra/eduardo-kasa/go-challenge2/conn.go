package gc2

import (
	"crypto/rand"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"code.google.com/p/go.crypto/nacl/box"
)

// Conn is a net.Conn using SecureReader and SecureWriter to transmit and receive
// encrypted content.
type Conn struct {
	conn     net.Conn
	isClient bool
	id       [32]byte

	// constant after handshake; protected by handshakeMutex
	handshakeMutex    sync.Mutex
	handshakeErr      error
	handshakeComplete bool

	// input/output
	sr io.Reader
	sw io.Writer
}

// LocalAddr returns the local network address.
func (c *Conn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (c *Conn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// SetDeadline sets the read and write deadlines associated with the connection.
// A zero value for t means Read and Write will not time out.
func (c *Conn) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

// SetReadDeadline sets the read deadline on the underlying connection.
// A zero value for t means Read will not time out.
func (c *Conn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

// SetWriteDeadline sets the write deadline on the underlying connection.
// A zero value for t means Write will not time out.
func (c *Conn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}

// Read reads data from the connection.
func (c *Conn) Read(p []byte) (n int, err error) {
	if err = c.Handshake(); err != nil {
		return 0, err
	}
	return c.sr.Read(p)
}

// Write writes data to the connection.
func (c *Conn) Write(p []byte) (n int, err error) {
	if err = c.Handshake(); err != nil {
		return 0, err
	}
	return c.sw.Write(p)
}

// Close closes the connection.
func (c *Conn) Close() error {
	c.handshakeMutex.Lock()
	defer c.handshakeMutex.Unlock()
	if err := c.conn.Close(); err != nil {
		return err
	}
	return nil
}

func (c *Conn) Handshake() error {
	c.handshakeMutex.Lock()
	defer c.handshakeMutex.Unlock()
	if err := c.handshakeErr; err != nil {
		return err
	}
	if c.handshakeComplete {
		return nil
	}

	if c.isClient {
		c.handshakeErr = c.clientHandshake()
	} else {
		c.handshakeErr = c.serverHandshake()

	}
	return c.handshakeErr
}

func (c *Conn) clientHandshake() error {
	cpk, csk, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}

	// key exchange
	_, err = c.conn.Write(cpk[:]) // send client public key to server
	if err != nil {
		return fmt.Errorf("sending public key %s", err)
	}
	spk := new([32]byte)
	_, err = io.ReadFull(c.conn, spk[:]) // recv server public key
	if err != nil {
		return fmt.Errorf("recv public key %s", err)
	}
	// end serverKey exchagne

	copy(c.id[:], spk[:])
	c.sr = NewSecureReader(c.conn, csk, spk)
	c.sw = NewSecureWriter(c.conn, csk, spk)

	c.handshakeComplete = true
	return nil
}

func (c *Conn) serverHandshake() error {
	spk, ssk, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}

	// key exchange
	cpk := new([32]byte)
	_, err = c.conn.Read(cpk[:]) // recv client public key
	if err != nil {
		return err
	}
	_, err = c.conn.Write(spk[:]) // send server public key to client
	if err != nil {
		return err
	}
	// end key exchange

	copy(c.id[:], cpk[:])
	c.sr = NewSecureReader(c.conn, ssk, cpk)
	c.sw = NewSecureWriter(c.conn, ssk, cpk)

	c.handshakeComplete = true
	return nil
}
