// Package messenger was created for the golang-challenge number 2
// (http://golang-challenge.com/go-challenge2/) by:
// Adrian Fletcher
// github.com/AdrianFletcher
// adrian@fletchtechnology.com.au

// Package messenger provides a basic server/client message relay service.
// Clients of the package can provide their own handshake implementation by
// providing a type conforming to the Handshaker interface.
package messenger

import (
	"io"
	"net"
)

// Handshake defines a basic handshake protocol (send and receive)
type Handshaker interface {
	SendHandshake(w io.Writer) (err error)
	GetHandshake(r io.Reader) (err error)
}

// Dial connects to the server, perform the handshake
// and return a connection.
func Dial(addr string, hs Handshaker) (net.Conn, error) {
	// Connect to server
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	if err = Handshake(conn, hs); err != nil {
		return nil, err
	}

	return conn, nil
}

// Serve starts an echo server on the given connection. When a new connection is
// establised, Server first checks for the handshake.
func Serve(conn net.Conn, hs Handshaker, w io.Writer, r io.Reader) error {
	defer conn.Close()
	if err := Handshake(conn, hs); err != nil {
		return err
	}

	buf := make([]byte, 32*1024)
	n, _ := r.Read(buf)
	n, _ = w.Write(buf[:n])

	return nil
}

// Handshake sends a handshake message then retrieves one in return.
func Handshake(conn net.Conn, hs Handshaker) error {
	if err := hs.SendHandshake(conn); err != nil {
		return err
	}
	if err := hs.GetHandshake(conn); err != nil {
		return err
	}
	return nil
}
