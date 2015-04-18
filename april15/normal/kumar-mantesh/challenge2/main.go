package main

import (
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"golang.org/x/crypto/nacl/box"
)

const maxMessageSize int = 32768 + box.Overhead
const keySize int = 32
const nonceSize int = 24

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	//generate priv and pub keys
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	//insecure communication to exchange public keys, handshake
	conn.Write(pub[:])
	serverPubB := make([]byte, keySize)
	n, err := conn.Read(serverPubB)
	if err != nil {
		return nil, err
	}
	if n != keySize {
		return nil, errors.New("Server didn't send the public key")
	}

	var serverPub [keySize]byte
	copy(serverPub[:], serverPubB)

	secureR := NewSecureReader(conn, priv, &serverPub)
	secureW := NewSecureWriter(conn, priv, &serverPub)

	return secureReadWriteCloser{secureR, secureW, conn}, nil
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	conn, err := l.Accept()
	defer conn.Close()
	if err != nil {
		return err
	}

	//The first request is assumed to be the client public key
	clientPubB := make([]byte, keySize)
	n, err := conn.Read(clientPubB)
	if err != nil {
		return err
	}
	if n != keySize {
		conn.Write([]byte("Public public key required"))
		return errors.New("Client didn't send the public key, terminating server")
	}
	var clientPub [keySize]byte
	copy(clientPub[:], clientPubB)

	//generate priv and pub keys for server
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}
	secureR := NewSecureReader(conn, priv, &clientPub)
	secureW := NewSecureWriter(conn, priv, &clientPub)

	//reply to the handshake with the public key
	_, err = conn.Write(pub[:])
	if err != nil {
		return err
	}

	b := make([]byte, maxMessageSize)
	n, err = secureR.Read(b)
	if err != nil {
		return err
	}
	b = b[:n]
	_, err = secureW.Write(b)
	if err != nil {
		return err
	}
	return nil
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
	conn.Close()
	fmt.Printf("%s\n", buf[:n])
}
