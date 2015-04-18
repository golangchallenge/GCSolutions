package gc2

import (
	"io"
	"log"
	"net"
)

func Server(conn net.Conn) *Conn {
	return &Conn{conn: conn, isClient: false}
}

// Room represents a very crude chat room.
type Room struct {
	clients map[[32]byte]*Conn
}

// Write writes p to all clients in the Room
func (r *Room) Write(p []byte) (n int, err error) {
	for _, client := range r.clients {
		go func(c *Conn) {
			_, err := c.Write(p)
			if err != nil {
				log.Printf("err writing to client (%s): %s", c.id, err)
			}
		}(client)
	}
	return len(p), nil
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	cr := &Room{clients: make(map[[32]byte]*Conn)}
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go handleConn(conn, cr)
	}
}

func handleConn(conn net.Conn, cr *Room) error {
	sc := Server(conn)
	defer sc.Close()
	sc.Handshake()

	// add to chat room
	cr.clients[sc.id] = sc
	defer delete(cr.clients, sc.id)

	// echo all
	n, err := io.Copy(cr, sc)
	if err != nil {
		log.Printf("io.Copy: %d %s", n, err)
	}
	return err
}
