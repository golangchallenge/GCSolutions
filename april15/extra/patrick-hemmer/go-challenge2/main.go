// This is a simple echo server & client using NaCl encryption.
//
// In server mode, it listens for incoming connections, and then reads off
// those connections. Each read is immediately written back out to the client
// with no processing.
//
// In client mode, a message is accepted as a command line argument. The
// message is sent to the server, and the client prints out anything received.
// A trailing newline is conditionally added to the output to ensure the shell
// prompt is at the beginning of the line after exit.
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
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

	// close the writer. The server will close the conn when it's done.
	if err := conn.Conn.(*net.TCPConn).CloseWrite(); err != nil {
		log.Fatal(err)
	}

	_, err = io.Copy(os.Stdout, conn)
	fmt.Printf("\n")
	if err != nil {
		log.Fatal(err)
	}
}
