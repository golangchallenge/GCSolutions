package main

import (
	"crypto/rand"
	"encoding/binary"
	"io"
	"net"

	"golang.org/x/crypto/nacl/box"
)

type conn struct {
	naclReader
	naclWriter
	raw net.Conn

	closed bool
}

func (c *conn) Close() error {
	if !c.closed {
		c.closed = true
		return c.raw.Close()
	}
	return nil
}

var _ io.ReadWriteCloser = &conn{}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	raw, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	rw, err := newConn(raw)
	if err != nil {
		raw.Close()
		return nil, err
	}

	return rw, nil

}

func newConn(raw net.Conn) (*conn, error) {
	mypub, mypriv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	c := &conn{
		raw: raw,
	}

	peersPub, err := c.doHandshake(*mypub)
	if err != nil {
		return nil, err
	}

	c.naclReader = *newSecureReader(c.raw, mypriv, &peersPub)
	c.naclWriter = *newSecureWriter(c.raw, mypriv, &peersPub)

	return c, nil

}

func (c *conn) doHandshake(mypub [keyLen]byte) (peersPub [keyLen]byte, err error) {
	errC := make(chan error)

	go func() {
		errC <- binWrite(c.raw, mypub)
	}()

	go func() {
		errC <- binRead(c.raw, &peersPub)
	}()

	for i := 0; i < 2; i++ {
		if err == nil {
			err = <-errC
		} else {
			<-errC
		}
	}

	return peersPub, err
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	for {
		rw, e := l.Accept()
		if e != nil {
			// TODO: e.Temporary()?
			return e
		}

		go func(rw net.Conn) {
			c, err := newConn(rw)
			if err != nil {
				rw.Close()
				return
			}

			Echo(c)
			c.Close()
		}(rw)

	}

}

func Echo(rw io.ReadWriteCloser) {
	io.Copy(rw, rw)
}

func binRead(r io.Reader, d interface{}) error {
	return binary.Read(r, binary.LittleEndian, d)
}

func binWrite(w io.Writer, d interface{}) error {
	return binary.Write(w, binary.LittleEndian, d)
}
