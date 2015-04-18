package main

import (
	"log"
	"net"
	"runtime"
)

type ConnHandler interface {
	HandleConn(conn net.Conn)
}

type ConnHandlerFunc func(conn net.Conn)

func (f ConnHandlerFunc) HandleConn(conn net.Conn) {
	f(conn)
}

// Simple, generic server. Creates a separate goroutine per connection.
type Server struct {
	Handler ConnHandler // Connection handler to invoke. Must be not nil.
}

// Accepts incoming connection on the listener and starts a new goroutine.
// The spawned goroutine calls srv.Handler for connection.
func (srv *Server) Serve(l net.Listener) error {
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go srv.serveConn(conn)
	}
	return nil
}

func (srv *Server) serveConn(conn net.Conn) {
	defer func() {
		// Isolate panics in handlers.
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			clientAddr := conn.RemoteAddr().String()
			log.Println("panic serving %v: %v\n%s", clientAddr, err, buf)
		}
		conn.Close()
	}()
	srv.Handler.HandleConn(conn)
}

type listener struct {
	net.Listener
}

// Returns box-encripted connection. Overwrites net.Listener's Accept method.
func (l *listener) Accept() (conn net.Conn, err error) {
	rawConn, err := l.Listener.Accept()
	if err != nil {
		return
	}
	conn = &Conn{conn: rawConn}
	return
}

// NewListener wraps given net.Listener as box-encripted listener.
// The resulting listener returns encripted connection from Accept method.
func NewListener(inner net.Listener) net.Listener {
	return &listener{inner}
}

// Dial connects to a tcp server, performs a handshake and
// returns a box-encripted connection.
func Dial(addr string) (conn *Conn, err error) {
	dialer := net.Dialer{}
	var rawConn net.Conn
	rawConn, err = dialer.Dial("tcp", addr)
	if err != nil {
		return
	}
	boxConn := &Conn{conn: rawConn, isClient: true}
	/*
		err = boxConn.Handshake()
		if err != nil {
			rawConn.Close()
			return
		}*/
	return boxConn, nil
}
