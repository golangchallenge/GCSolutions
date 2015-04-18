package main

import (
	"fmt"
	"io"
	"net"
)

// Server is the secure echo server.
type Server struct {
	keyPair *KeyPair
}

// NewServer initializes a new Server with its own keys. The server will
// perform a handshake with each client to exchange public keys.
func NewServer(kp *KeyPair) *Server {
	return &Server{kp}
}

// Serve starts an infinite loop waiting for client connections.
func (s *Server) Serve(l net.Listener) error {
	for {
		conn, err := l.Accept()
		if err != nil {
			s.debug("Failed to accept client: %s\n", err)
			return err
		}
		go func(conn net.Conn) {
			defer conn.Close()
			commonKey, err := s.handshake(conn)
			if err != nil {
				s.debug("Error performing handshake: %s\n", err)
			}
			if err := s.handle(conn, commonKey); err != nil {
				s.debug("Error handling client: %s\n", err)
			}
		}(conn)
	}
}

// handshake performs the key exchange with the client, returning the shared
// key that can be used to communicate with that client only.
func (s *Server) handshake(conn io.ReadWriter) (*[keySize]byte, error) {
	s.debug("Performing key exchange...\n")
	kp, err := s.keyPair.Exchange(conn)
	if err != nil {
		return nil, err
	}
	commonKey := kp.CommonKey()
	return commonKey, nil
}

// handle takes care of client/server behavior after the handshake.
func (s *Server) handle(conn io.ReadWriter, commonKey *[keySize]byte) error {
	// Setup encrypted reader/writer to communicate with the client.
	sr := &SecureReader{conn, commonKey}
	sw := &SecureWriter{conn, commonKey}

	// Read decrypted data from the client.
	s.debug("Reading...\n")
	buf := make([]byte, maxMessageSize)
	c, err := sr.Read(buf)
	if err != nil {
		return err
	}
	s.debug("Read %d bytes: %s\n", c, buf[:c])

	// Write encrypted data back to the client.
	s.debug("Writing...\n")
	c, err = sw.Write(buf[:c])
	if err != nil {
		return err
	}
	s.debug("Wrote %d bytes\n", c)

	return nil
}

func (s *Server) debug(str string, v ...interface{}) {
	debugf("server: %s", fmt.Sprintf(str, v...))
}

// Client is the secure echo client.
type Client struct {
	keyPair   *KeyPair
	commonKey *[32]byte
}

// NewClient initializes a Client with its own keys. The client will perform a
// handshake with the server to exchange public keys.
func NewClient(kp *KeyPair) *Client {
	return &Client{kp, nil}
}

// Handshake performs the public key exchange with the server.
func (c *Client) Handshake(conn io.ReadWriter) error {
	c.debug("Performing key exchange...\n")
	kp, err := c.keyPair.Exchange(conn)
	if err != nil {
		return err
	}
	c.commonKey = kp.CommonKey()
	return nil
}

// SecureConn returns a ReadWriteCloser to communicate with the server.
// Requires that the shared key has been provided, probably by getting it via
// Handshake.
func (c *Client) SecureConn(conn io.ReadWriteCloser) io.ReadWriteCloser {
	r := &SecureReader{conn, c.commonKey}
	w := &SecureWriter{conn, c.commonKey}
	return struct {
		io.Reader
		io.Writer
		io.Closer
	}{r, w, conn}
}

func (c *Client) debug(str string, v ...interface{}) {
	debugf("client: %s", fmt.Sprintf(str, v...))
}
