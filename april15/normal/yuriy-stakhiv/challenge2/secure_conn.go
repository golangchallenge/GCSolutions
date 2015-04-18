package main

import (
	"fmt"
	"net"
)

// SecureConn wraps net.Conn
type SecureConn struct {
	conn net.Conn
	rw   *SecureReadWriter
}

// NewSecureConn creates new SecureConn
func NewSecureConn(conn net.Conn) *SecureConn {
	sc := &SecureConn{}
	sc.conn = conn
	return sc
}

// Close underlaying connection
func (sc *SecureConn) Close() error {
	return sc.conn.Close()
}

func (sc *SecureConn) Read(b []byte) (int, error) {
	return sc.rw.Read(b)
}

func (sc *SecureConn) Write(b []byte) (int, error) {
	return sc.rw.Write(b)
}

// ClientHandshake implements client handshake part
func (sc *SecureConn) ClientHandshake(pub, priv *[32]byte) error {
	var rPub [32]byte
	n, err := sc.conn.Read(rPub[:])
	if err != nil {
		return err
	}
	if n < len(rPub) {
		return fmt.Errorf("bad remote key")
	}

	_, err = sc.conn.Write(pub[:])
	if err != nil {
		return err
	}
	sc.rw = NewSecureReadWriter(priv, &rPub, sc.conn, sc.conn)
	return nil
}

// ServerHandshake implements server handshake part
func (sc *SecureConn) ServerHandshake(pub, priv *[32]byte) error {
	_, err := sc.conn.Write(pub[:])
	if err != nil {
		return err
	}

	var rPub [32]byte
	n, err := sc.conn.Read(rPub[:])
	if err != nil {
		return err
	}
	if n < len(rPub) {
		return fmt.Errorf("bad client remote key")
	}

	sc.rw = NewSecureReadWriter(priv, &rPub, sc.conn, sc.conn)
	return nil
}
