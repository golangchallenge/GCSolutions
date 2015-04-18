package main

import (
	"io"
	"net"
)

// SecureConn wrapps net.Conn with a secure reader/writer.
type SecureConn struct {
	conn net.Conn
	io.Reader
	io.Writer
}

// Close closes the connection.
func (s *SecureConn) Close() error {
	return s.conn.Close()
}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	pub, priv, err := generateKey()
	if err != nil {
		return nil, err
	}
	c, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	// Handshake, exchange public key with server
	_, err = c.Write(pub[:])
	if err != nil {
		return nil, err
	}

	serverPubKey := new([32]byte)
	_, err = c.Read(serverPubKey[:])
	if err != nil {
		return nil, err
	}
	conn := &SecureConn{c, NewSecureReader(c, priv, serverPubKey),
		NewSecureWriter(c, priv, serverPubKey)}
	return conn, nil
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	for {
		pub, priv, err := generateKey()
		if err != nil {
			return err
		}
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go func(c net.Conn) {
			defer c.Close()
			clientPubKey := new([32]byte)
			_, err := c.Read(clientPubKey[:])
			if err != nil {
				return
			}
			_, err = c.Write(pub[:])
			if err != nil {
				return
			}
			conn := &SecureConn{c, NewSecureReader(c, priv, clientPubKey),
				NewSecureWriter(c, priv, clientPubKey)}
			buf := make([]byte, MaxMsgLength)
			n, err := conn.Read(buf)
			if err != nil {
				return
			}
			_, err = conn.Write(buf[:n])
			if err != nil {
				return
			}
		}(conn)
	}
}
