package main

import (
	"io"
	"net"
	"sync"
	"time"
)

// A boxed connection. Implemets the net.Conn interface.
// The connection is supposed to be secure from evesdropping;
// but is **not secure** from MITM.
type Conn struct {
	conn     net.Conn
	isClient bool

	handshakeMutex sync.Mutex
	handshakeDone  bool  // flag if handshake was done
	handshakeErr   error // error from handshake

	r boxReader // decoder of incoming data over connection
	w boxWriter // encoder of outcoming data over connection
}

// Returns the local network address.
func (c *Conn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

// Returns the remote network address.
func (c *Conn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// See net.Conn documentation.
func (c *Conn) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

// See net.Conn documentation.
func (c *Conn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

// See net.Conn documentation.
func (c *Conn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}

func (c *Conn) Read(b []byte) (n int, err error) {
	if err = c.Handshake(); err != nil {
		return
	}
	return c.r.Read(b)
}

func (c *Conn) Write(b []byte) (n int, err error) {
	if err = c.Handshake(); err != nil {
		return
	}
	return c.w.Write(b)
}

// Implementation of io.WriterTo. This allows io.Copy not to block on buffers.
func (c *Conn) WriteTo(w io.Writer) (n int64, err error) {
	if err = c.Handshake(); err != nil {
		return
	}
	return c.r.WriteTo(w)
}

func (c *Conn) Close() error {
	return c.conn.Close()
}

func (c *Conn) Handshake() error {
	c.handshakeMutex.Lock()
	defer c.handshakeMutex.Unlock()

	if c.handshakeErr == nil && !c.handshakeDone {
		err := c.doHandshake()
		c.handshakeErr = err
		c.handshakeDone = err == nil
	}
	return c.handshakeErr
}

func (c *Conn) doHandshake() (err error) {
	rPub, rPriv, err := GenerateKey()
	if err != nil {
		return
	}
	wPub, wPriv, err := GenerateKey()
	if err != nil {
		return
	}

	if c.isClient {
		err = c.clientHandshake(rPub, wPub)
	} else {
		err = c.serverHandshake(rPub, wPub)
	}
	if err != nil {
		return
	}
	// Reuse memory of private keys for shared keys.
	c.r.set(c.conn, MakeSharedKey(rPriv, rPub, rPriv))
	c.w.set(c.conn, MakeSharedKey(wPriv, wPub, wPriv))
	return
}

func (c *Conn) clientHandshake(wPub, rPub *[32]byte) (err error) {
	err = c.sendKeys(wPub, rPub)
	if err != nil {
		return
	}
	return c.recvKeys(rPub, wPub)
}

func (c *Conn) serverHandshake(wPub, rPub *[32]byte) (err error) {
	var rPeer, wPeer [32]byte
	err = c.recvKeys(&rPeer, &wPeer)
	if err != nil {
		return
	}
	err = c.sendKeys(wPub, rPub)
	if err != nil {
		return
	}
	*wPub = wPeer
	*rPub = rPeer
	return
}

func (c *Conn) sendKeys(wPub, rPub *[32]byte) (err error) {
	_, err = c.conn.Write(wPub[:])
	if err != nil {
		return
	}
	_, err = c.conn.Write(rPub[:])
	return
}

func (c *Conn) recvKeys(rPeer, wPeer *[32]byte) (err error) {
	_, err = io.ReadFull(c.conn, rPeer[:])
	if err != nil {
		return
	}
	_, err = io.ReadFull(c.conn, wPeer[:])
	return
}
