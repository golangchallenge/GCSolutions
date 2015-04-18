// Package main implements a method of communicating securely over a network. It provides
// a secure echo server as an example implementation.
package main

import (
	"crypto/rand"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"golang.org/x/crypto/nacl/box"
)

// Dial generates a private/public key pair,
// connects to the server, performs the handshake
// and returns a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	// Receive the server's public key
	var peersPub [32]byte
	err = binary.Read(conn, binary.BigEndian, &peersPub)
	if err != nil {
		return nil, err
	}

	// Send our public key
	err = binary.Write(conn, binary.BigEndian, pub)
	if err != nil {
		return nil, err
	}

	// Create Secured Reader/Writer
	secureConn := NewSecureConnection(conn, priv, &peersPub)

	return secureConn, nil
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	for {
		// Block until a connection is received
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		// Process in a new goroutine, so multiple requests may be processed.
		go func(c net.Conn) {
			defer c.Close()

			pub, priv, err := box.GenerateKey(rand.Reader)
			if err != nil {
				fmt.Println("Error generating keys:", err)
				return
			}

			// Send server's public key in plaintext
			err = binary.Write(conn, binary.BigEndian, pub)
			if err != nil {
				fmt.Println("Error writing public key:", err)
				return
			}

			// Receive the client's public key
			var peersPub [32]byte
			binary.Read(conn, binary.BigEndian, &peersPub)

			secureConn := NewSecureConnection(conn, priv, &peersPub)

			buf := make([]byte, 2048)
			n, err := secureConn.Read(buf)
			if err != nil {
				fmt.Println("Error reading from connection:", err)
				return
			}

			_, err = secureConn.Write(buf[:n])
			if err != nil {
				fmt.Println("Error writing to connection:", err)
				return
			}
		}(conn)
	}
}

func main() {
	port := flag.Int("l", 0, "Listen mode. Specify port")
	flag.Parse()

	// Server mode
	if *port != 0 {
		l, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
		if err != nil {
			log.Fatal(err)
		}
		defer l.Close()
		log.Fatal(Serve(l))
	}

	// Client mode
	if len(os.Args) != 3 {
		log.Fatalf("Usage: %s <port> <message>", os.Args[0])
	}
	conn, err := Dial("localhost:" + os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	if _, err := conn.Write([]byte(os.Args[2])); err != nil {
		log.Fatal(err)
	}
	buf := make([]byte, len(os.Args[2]))
	n, err := conn.Read(buf)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", buf[:n])
}
