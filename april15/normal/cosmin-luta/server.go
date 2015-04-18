//
// Go Challenge 2
//
// Cosmin Luță <q4break@gmail.com>
//
// server.go - Code for handling clients
//

package main

import "net"

func readAndEcho(c *EncryptedConnection, buf []byte) error {
	n, err := c.Read(buf)
	if err != nil {
		return err
	}

	if n > 0 {
		// echo back anything received from the client
		_, err = c.Write(buf[:n])
		if err != nil {
			return err
		}
	}
	return nil
}

// Handle a client connected to the server
func clientHandler(c net.Conn) error {
	defer c.Close()

	// Set up the encrypted connection
	encConn, err := NewEncryptedConnection(c)
	if err != nil {
		return err
	}

	buf := make([]byte, MaxMessageSize)

	// attempt to read as long as the client doesn't close the connection
	for {
		if err = readAndEcho(encConn, buf); err != nil {
			return err
		}
	}
}

// Serve loops in an infinite loop, accepting client connections.
func Serve(l net.Listener) error {
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go clientHandler(conn)
	}
}
