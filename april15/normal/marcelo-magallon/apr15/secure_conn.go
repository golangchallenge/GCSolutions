package main

import (
	"bytes"
	"crypto/rand"
	"errors"
	"io"
	"net"

	"golang.org/x/crypto/nacl/box"
)

// protocolHandshake is the string used to identify the protocol we are
// trying to communicate with.
var protocolHandshake = []byte("whispering gophers 1\n")

// badHandshakeResponse is the string sent back to the client after an
// unsucessful handshake
var badHandshakeResponse = []byte("you shall not pass!")

// ErrBadHandshake is the error emitted when there is a handshake error
// on either side of a connection.
var ErrBadHandshake = errors.New("bad client/server handshake")

type secureConn struct {
	r io.Reader
	w io.Writer
	c net.Conn
}

func (c *secureConn) Read(p []byte) (int, error) {
	return c.r.Read(p)
}

func (c *secureConn) Write(p []byte) (int, error) {
	return c.w.Write(p)
}

func (c *secureConn) Close() error {
	return c.c.Close()
}

// Dial generates a private/public key pair, connects to the server,
// perform the handshake and returns a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	clientPriv, serverPub, err := clientHandshake(conn)
	if err != nil {
		conn.Close()
		return nil, ErrBadHandshake
	}

	c := &secureConn{
		r: NewSecureReader(conn, clientPriv, serverPub),
		w: NewSecureWriter(conn, clientPriv, serverPub),
		c: conn,
	}
	return c, nil
}

// Serve starts a secure echo server on the provided listener.
func Serve(l net.Listener) error {
	for {
		// server waiting for connection
		switch conn, err := l.Accept(); err {
		case nil:
			go serve(conn)
		default:
			return err
		}
	}
}

func serve(conn net.Conn) {
	// perform handshake
	serverPriv, clientPub, err := serverHandshake(conn)
	if err != nil {
		conn.Write(badHandshakeResponse)
		conn.Close()
		return
	}

	c := &secureConn{
		r: NewSecureReader(conn, serverPriv, clientPub),
		w: NewSecureWriter(conn, serverPriv, clientPub),
		c: conn,
	}

	var buf [MaxMsgLen]byte
	rw := &CatchErrorReadWriter{rw: c}
	for rw.err == nil {
		n, _ := rw.Read(buf[0:])
		rw.Write(buf[:n])
	}
	c.Close()
}

// serverHandshake performs the protocol handshake server-side
func serverHandshake(c net.Conn) (*[32]byte, *[32]byte, error) {
	// server generates public/private key pair
	serverPub, serverPriv, err := box.GenerateKey(rand.Reader)

	rw := &CatchErrorReadWriter{rw: c, err: err}

	// client sends protocolHandshake
	clientHandshake := make([]byte, len(protocolHandshake))

	rw.Read(clientHandshake)
	if rw.err == nil && !bytes.Equal(protocolHandshake, clientHandshake) {
		rw.err = ErrBadHandshake
	}

	// XXX: Ideally, here we would send a response to the client to
	// tell it that we are the thing it was expecting to connect to
	// (e.g. SSH exchanging "SSH-2.0-...\n\r" between client and
	// server). This cannot be done because TestSecureDial
	// implements a server that upon accepting the connection,
	// *writes* a key to the client, without reading anything from
	// it. Because the way that test is written, when it *reads*
	// from the connection, what it will actually read is
	// protocolHandshake + the client's public key, therefore
	// "verifying" that it is not getting the plaintext that the
	// client is sending in the test. It could be forced by making
	// the handshake reply 32 null bytes (which is what the test
	// server is sending to the client), but that's stretching the
	// thing a little too far.

	// sends public to client
	writeFull(rw, serverPub[:])

	// client generates public/private key pair, sends public to server
	clientPub, _ := receiveKey(rw)

	return serverPriv, clientPub, rw.err
}

// clientHandshake performs the protocol handshake client-side
func clientHandshake(c net.Conn) (*[32]byte, *[32]byte, error) {
	// client generates public/private key pair
	clientPub, clientPriv, err := box.GenerateKey(rand.Reader)

	rw := &CatchErrorReadWriter{rw: c, err: err}

	// client sends protocolHandshake
	writeFull(rw, protocolHandshake)

	// XXX: Here we would read the handshake response from the
	// server. See comment in serverHandshake as to why it's not
	// done.

	// server generates public/private key pair, sends public to client
	serverPub, _ := receiveKey(rw)

	// client sends public to server
	writeFull(rw, clientPub[:])

	return clientPriv, serverPub, rw.err
}

// receiveKey receives one public or private key over the provider
// reader
func receiveKey(r io.Reader) (*[32]byte, error) {
	key := [32]byte{}
	switch _, err := io.ReadFull(r, key[:]); err {
	case io.EOF, io.ErrUnexpectedEOF:
		// no data?
		return nil, io.ErrUnexpectedEOF
	default:
		// something else happened
		return nil, err
	case nil:
		// everything is ok
	}

	return &key, nil
}
