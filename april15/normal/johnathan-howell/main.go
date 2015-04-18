package main

import (
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"golang.org/x/crypto/nacl/box"
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
	// Generate a key pair and send our public key to the server
	clientpub, priv, err := box.GenerateKey(rand.Reader)
	_, err = conn.Write(clientpub[:])
	if err != nil {
		return nil, err
	}
	// Recieve the server's public key
	var serverpub [32]byte
	_, err = conn.Read(serverpub[:])
	if err != nil {
		return nil, err
	}

	secureW := NewSecureWriter(conn, priv, &serverpub)
	secureR := NewSecureReader(conn, priv, &serverpub)
	return NewSecureCodec(secureR, secureW, conn), nil
}

// handleNewConnection handles a connection to the listen server and writes any errors to errs
func handleNewConnection(conn net.Conn, errs chan error) {
	defer conn.Close()
	// Recieve the client's public key
	rcvbuf := make([]byte, 1024)
	n, err := conn.Read(rcvbuf)
	if err != nil {
		errs <- err
		return
	}
	// n, key length, should always be 32.
	if n != 32 {
		errs <- errors.New("WARNING: invalid keylen")
	}
	var clientpub [32]byte
	copy(clientpub[:], rcvbuf[:32])

	// Generate the server's keypair and send the public key to the client
	serverpub, priv, err := box.GenerateKey(rand.Reader)
	_, err = conn.Write(serverpub[:])
	if err != nil {
		errs <- err
		return
	}
	// Create a new SecureCodec to handle reading and writing encrypted data from client
	secureW := NewSecureWriter(conn, priv, &clientpub)
	secureR := NewSecureReader(conn, priv, &clientpub)
	secureCodec := NewSecureCodec(secureR, secureW, conn)

	// Read data from client, then echo it back
	n, err = secureCodec.Read(rcvbuf)
	if err != nil {
		errs <- err
		return
	}
	_, err = secureCodec.Write(rcvbuf[:n])
	if err != nil {
		errs <- err
		return
	}
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	errs := make(chan error)
	// This goroutine prints any errors encountered by handleNewConnection()
	go func() {
		for {
			select {
			case e := <-errs:
				log.Printf("Listen server reported an error: %s\n", e)
			}
		}
	}()
	// Start our connection listener, asynchronously handle connections using handleNewConnection
	var err error
	for {
		conn, err := l.Accept()
		if err != nil {
			break
		}
		go handleNewConnection(conn, errs)
	}
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
