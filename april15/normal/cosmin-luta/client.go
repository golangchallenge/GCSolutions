//
// Go Challenge 2
//
// Cosmin Luță <q4break@gmail.com>
//
// client.go - Code for establishing an encrypted connection
//

package main

import (
	"io"
	"net"
)

// Dial generates a private/public key pair, connects to the server,
// performs the handshake and returns a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {

	// Create a plain "tcp" connection to addr...
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	// ...and add encryption on top of it
	return NewEncryptedConnection(conn)
}
