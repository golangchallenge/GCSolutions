package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"./sio"
)

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	return sio.NewSecureReader(r, priv, pub)
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	return sio.NewSecureWriter(w, priv, pub)
}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (rwc io.ReadWriteCloser, err error) {
	var (
		conn    net.Conn
		conduit *sio.Conduit
	)

	if conn, err = net.Dial("tcp", addr); err != nil {
		return nil, err
	}

	if conduit, err = sio.NewConduit(conn); err != nil {
		return nil, err
	}

	return conduit, nil
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) (err error) {
	for {
		var conn net.Conn

		if conn, err = l.Accept(); err != nil {
			return
		}

		go func(c net.Conn) {
			var conduit *sio.Conduit

			defer c.Close()

			if conduit, err = sio.NewConduit(c); err != nil {
				log.Fatal(err)
			}

			// This is just an echo server, so... echo!
			if _, err = io.Copy(conduit, conduit); err != nil {
				log.Fatal(err)
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
