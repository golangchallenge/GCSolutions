// This is a multi-message secure echo server & client.
//
// In server mode, it listens for incoming connections, and then listens for
// messages on those connections. When a message is received, it is broadcasted
// out to all other connected clients.
//
// The client has 2 modes of operation:
//  * Single message mode - The client takes a message as an argument, sends
//    it to the server, and dumps all data from the server to STDOUT.
//  * Multi-message mode - The client reads complete lines from STDIN, sends
//    them to the server, and dumps all data from the server to STDOUT.
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	kp, err := NewKeyPair()
	if err != nil {
		return err
	}

	l = NewSecureListener(l, kp)

	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go serveEcho(conn)
	}
}

func serveEcho(conn net.Conn) error {
	_, err := io.Copy(conn, conn)
	conn.Close() // try to close, but ignore error
	return err
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
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <port> [message]", os.Args[0])
	}
	conn, err := Dial("localhost:" + os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	if len(os.Args) > 2 {
		// single message mode
		if _, err := conn.Write([]byte(strings.Join(os.Args[2:], " "))); err != nil {
			log.Fatal(err)
		}
		if err := conn.Conn.(*net.TCPConn).CloseWrite(); err != nil {
			log.Fatal(err)
		}
	} else {
		// multi-message mode
		go func() {
			if _, err := io.Copy(conn, os.Stdin); err != nil {
				log.Fatal(err)
			}
			conn.Conn.(*net.TCPConn).CloseWrite()
		}()
	}

	_, err = io.Copy(os.Stdout, conn)
	if err != nil {
		log.Fatal(err)
	}
}
