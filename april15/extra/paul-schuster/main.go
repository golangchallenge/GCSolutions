package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	return NewSecureConnection(conn)
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	s := NewServer()
	s.Serve(l, NewSecureConnection)
	return nil
}

func main() {
	client := flag.Int("i", 0, "Interactive Mode, Specify port")
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

	// Interactive Client mode
	if *client != 0 {
		conn, err := Dial(fmt.Sprintf("localhost:%d", *client))
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Golang Challenge 2 - extra Interactive UI\n")
		fmt.Printf("Connected to: localhost:%d\n", *client)
		go clientRead(conn)
		clientWrite(conn)
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
