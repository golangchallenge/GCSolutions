package main

import (
	"crypto/rand"
	"io"
	"net"
	"time"

	"golang.org/x/crypto/nacl/box"
)

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	s, err := newEchoServer()
	if err != nil {
		return err
	}

	var tempDelay time.Duration

	for {
		conn, err := l.Accept()
		if err != nil {
			if err, ok := err.(net.Error); ok && err.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}

				v(1).Printf("accept error: %v, retrying in %v", err, tempDelay)
				time.Sleep(tempDelay)

				continue
			}
			return err
		}

		tempDelay = 0

		v(2).Printf("server(%v): accepted connection", conn.RemoteAddr())

		c := s.newConn(conn)
		go c.serve()
	}
}

// echoServer echoes back data after unboxing it. It also
// is responsible for exchanging the keys with the client.
type echoServer struct {
	priv, pub *key
}

func newEchoServer() (*echoServer, error) {
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	pubK, privK := (*key)(pub), (*key)(priv)

	v(2).Printf("server: generated private key %v, public key %v", privK, pubK)

	s := &echoServer{
		priv: privK,
		pub:  pubK,
	}

	return s, nil
}

func (s *echoServer) newConn(conn net.Conn) *serverConn {
	return &serverConn{
		s:    s,
		conn: conn,
	}
}

type serverConn struct {
	s *echoServer

	conn net.Conn
}

func (c *serverConn) serve() {
	defer func() {
		v(2).Printf("server(%v): closing conn", c.conn.RemoteAddr())
		c.conn.Close()
	}()

	peersPub, err := exchange(c.conn, c.s.pub)
	if err != nil {
		v(1).Printf("server(%v): could not exchange keys: %v", c.conn.RemoteAddr(), err)
		return
	}

	v(2).Printf("server(%v): peer's public key is %v", c.conn.RemoteAddr(), peersPub)

	securedR := NewSecureReader(c.conn, c.s.priv, peersPub)
	securedW := NewSecureWriter(c.conn, c.s.priv, peersPub)

	// We can afford to use io.Copy to ferry data from reader to writer because the challenge
	// rules say that the upper limit to the message size is 32kb.
	if _, err := io.Copy(securedW, securedR); err != nil {
		v(1).Printf("server(%v): could not echo data back: %v", c.conn.RemoteAddr(), err)
	}
}
