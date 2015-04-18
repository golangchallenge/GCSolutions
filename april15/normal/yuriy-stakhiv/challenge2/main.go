package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"golang.org/x/crypto/nacl/box"
	"io"
	"log"
	"net"
	"os"
)

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	return NewSecureReadWriter(priv, pub, r, nil)
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	return NewSecureReadWriter(priv, pub, nil, w)
}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	sConn := NewSecureConn(conn)
	err = sConn.ClientHandshake(pub, priv)
	if err != nil {
		return nil, err
	}
	return sConn, nil
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	server, err := NewEchoServer()
	if err != nil {
		return err
	}
	return server.Serve(l)
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
