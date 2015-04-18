package secure

import (
	"crypto/rand"
	"log"
	"net"

	"golang.org/x/crypto/nacl/box"
)

// Server handles incoming NaCl encrypted connections and echoes messages back to clients.
type Server struct {
	l          net.Listener
	privateKey *[32]byte
	publicKey  *[32]byte
}

// NewServer instantiates a new Server
func NewServer(l net.Listener) (server *Server, err error) {
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	return &Server{
		l:          l,
		privateKey: priv,
		publicKey:  pub,
	}, nil
}

// Serve starts the server loop.
func (serv *Server) Serve() (err error) {
	for {
		conn, err := serv.l.Accept()
		if err != nil {
			return err
		}
		go serv.handleConnection(conn)
	}
}

func (serv *Server) handleConnection(conn net.Conn) {
	secureConn, err := NewConn(conn)
	if err != nil {
		conn.Close()
		log.Printf("SERVER: %s", err)
		return
	}
	defer secureConn.Close()

	msg := make([]byte, 2048)
	n, err := secureConn.Read(msg)
	if err != nil {
		log.Printf("SERVER: %s", err)
		return
	}
	msg = msg[:n]

	if _, err := secureConn.Write(msg); err != nil {
		log.Printf("SERVER: %s", err)
		return
	}
}
