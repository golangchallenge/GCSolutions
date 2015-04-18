package main

import (
	"io"
	"net"
)

type EncryptedConnection struct {
	conn net.Conn
	sw   io.Writer
	sr   io.Reader
}

func NewEncryptedConnection(conn net.Conn, priv, pub *[32]byte) io.ReadWriteCloser {
	sw := NewSecureWriter(conn, priv, pub)
	sr := NewSecureReader(conn, priv, pub)
	return &EncryptedConnection{conn, sw, sr}
}

func (ec *EncryptedConnection) Read(out []byte) (int, error) {
	return ec.sr.Read(out)
}

func (ec *EncryptedConnection) Write(message []byte) (int, error) {
	return ec.sw.Write(message)
}

func (ec *EncryptedConnection) Close() error {
	return ec.conn.Close()
}
