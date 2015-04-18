package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

// Dial creates a SecureConnection
func Dial(addr string) (io.ReadWriteCloser, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	return NewSecureConnection(conn)

}

// handleConnection processes exactly one client request.
func handleConnection(conn net.Conn) {
	buf := make([]byte, 32*1024)
	defer conn.Close()

	secureConn, err := NewSecureConnection(conn)
	if err != nil {
		log.Fatal(err)
	}

	n, err := secureConn.Read(buf)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := secureConn.Write(buf[:n]); err != nil {
		log.Fatal(err)
	}
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go handleConnection(conn)
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
