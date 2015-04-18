package main

import (
	"io"
	"log"
	"net"
)

// peer holds the information on a connection.
type peer struct {
	id        int
	send      chan []byte
	broadcast chan []byte
	conn      io.ReadWriteCloser
}

// NewPeer creates a Peer that can send and revieve messages in a server.
func NewPeer(s *Server, conn io.ReadWriteCloser) {
	p := peer{}
	p.id = s.nextID()
	p.send = make(chan []byte)
	p.broadcast = s.broadcast
	p.conn = conn

	s.addPeer <- p
	go p.outHandler()
	p.inHandler()
	s.rmPeer <- p
}

// outHandler sends all the broadcast info to a peer.
func (p *peer) outHandler() {
	for {
		select {
		case m, more := <-p.send:
			if !more {
				break
			}
			_, err := p.conn.Write(m)
			if err != nil {
				log.Print(err)
				break
			}
		}
	}
}

// inHandler broadcasts all info from a peer.
func (p *peer) inHandler() {
	buf := make([]byte, 32000)
	for {
		n, err := p.conn.Read(buf)
		if err != nil {
			log.Print(err)
			break
		}
		if n > 0 {
			p.broadcast <- buf[:n]
		}
	}
}

// Server represents a simple chat server.
type Server struct {
	peers     map[int]peer
	usedIds   map[int]bool
	addPeer   chan peer
	rmPeer    chan peer
	broadcast chan []byte
}

// NewServer creates and setups a server.
func NewServer() Server {
	s := Server{}
	s.peers = make(map[int]peer)
	s.usedIds = make(map[int]bool)
	s.addPeer = make(chan peer)
	s.rmPeer = make(chan peer)
	s.broadcast = make(chan []byte)
	return s
}

// Serve starts a chat server on a given net.Listener, with a connHandler middleware for the io stream.
func (s *Server) Serve(l net.Listener, connHandler func(net.Conn) (io.ReadWriteCloser, error)) error {
	go s.dispatcher()
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		c, err := connHandler(conn)
		if err != nil {
			return err
		}
		go NewPeer(s, c)
	}
}

// dispatcher shunts the data around all the peers and broadcast.
func (s *Server) dispatcher() {
	for {
		select {
		case a := <-s.addPeer:
			s.peers[a.id] = a
		case d := <-s.rmPeer:
			close(d.send)
			delete(s.peers, d.id)
			delete(s.usedIds, d.id)
		case b := <-s.broadcast:
			for _, v := range s.peers {
				v.send <- b
			}
		}
	}
}

// nextID returns the next unused ID num, note it will reuse IDs that are no longer in use.
func (s *Server) nextID() int {
	for i := 0; i <= len(s.usedIds); i++ {
		if ok, _ := s.usedIds[i]; !ok {
			s.usedIds[i] = true
			return i
		}
	}
	s.usedIds[len(s.usedIds)] = true
	return len(s.usedIds)
}
