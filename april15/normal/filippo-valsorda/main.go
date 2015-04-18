package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

// Dial connects to the server, perform the handshake and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	// Open a new connection
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	// And then wrap it in a SecureConn
	return NewSecureConn(conn)
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	defer l.Close()
	for {
		// Accept a new connection
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		go func(conn net.Conn) {
			// Wrap it in a SecureConn
			c, err := NewSecureConn(conn)
			if err != nil {
				log.Print(err)
				return
			}

			// And echo back
			if _, err := io.Copy(c, c); err != nil {
				log.Print(err)
			}
			c.Close()
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
	if err != nil && err != io.EOF {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", buf[:n])
}
