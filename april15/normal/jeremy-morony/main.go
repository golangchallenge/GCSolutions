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
	c, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	return establishSecureConn(c)
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		go func(c net.Conn) {
			defer c.Close()

			secure, err := establishSecureConn(c)
			if err != nil {
				return
			}

			buf := make([]byte, 32*1024)
			for {
				n, err := secure.Read(buf)
				if err != nil && err != io.EOF {
					break
				}
				secure.Write(buf[:n])
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
	if err != nil && err != io.EOF {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", buf[:n])
}
