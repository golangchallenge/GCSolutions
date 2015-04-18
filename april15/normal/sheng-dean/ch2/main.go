// Package main implements readers and writers that utilizes NaCL to securely
// encrypt/decrypt messages.
package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"golang.org/x/crypto/nacl/box"
)

// Dial generates a private/public key pair and connects to the server
// located at the provided address. Once connected, it performs a handshake
// by exchanging public keys. It returns a SecureReadWriteCloser if
// the handshake is successful as well as any errors encountered.
func Dial(addr string) (io.ReadWriteCloser, error) {
	clientPub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	// begin handshake
	serverPub := &[32]byte{}
	rw := newErrReadWriter(conn)
	rw.write(clientPub[:])
	rw.read(serverPub[:])
	if rw.err != nil {
		return nil, rw.err
	}

	return newSecureConnection(conn, priv, serverPub), nil
}

// Serve starts a secure echo server on the given listener.
// Clients must be prepared to perform a handshake by exchanging a 32 byte
// public key upon connecting.
func Serve(l net.Listener) error {
	serverPub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		defer conn.Close()
		go echo(conn, serverPub, priv)
	}
}

func echo(conn net.Conn, serverPub, priv *[32]byte) (int, error) {
	// begin handshake
	rw := newErrReadWriter(conn)
	rw.write(serverPub[:])
	clientPub := &[32]byte{}
	rw.read(clientPub[:])
	if rw.err != nil {
		return rw.n, rw.err
	}

	secure := newErrReadWriter(newSecureConnection(conn, priv, clientPub))

	for {
		// echo messages. Hint 1 says messages will be at most 32kb
		buf := make([]byte, 32768)
		n := secure.read(buf)
		secure.write(buf[:n])
		if secure.err != nil {
			return secure.n, secure.err
		}
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
