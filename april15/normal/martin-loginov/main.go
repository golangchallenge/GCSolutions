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

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	//Generate a keypair for the client
	cpub, cpriv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	//Perform the handshake with the server
	//Receive the public key of the server
	var spub [32]byte
	_, err = conn.Read(spub[:])
	if err != nil {
		return nil, err
	}

	//Send our public key to the server
	_, err = conn.Write(cpub[:])
	if err != nil {
		return nil, err
	}

	//Initialise the secure socket with the correct keys
	secread := NewSecureReader(conn, cpriv, &spub)
	secwrite := NewSecureWriter(conn, cpriv, &spub)

	return SecureSocket{
		Reader: secread,
		Writer: secwrite,
		Closer: conn,
	}, err
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	conn, err := l.Accept()
	if err != nil {
		return err
	}

	//Generate private and public keys for the server
	spub, spriv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}

	//Perform the handshake with the client
	//Public key of the client
	var cpub [32]byte

	_, err = conn.Write(spub[:])
	if err != nil {
		return err
	}

	_, err = conn.Read(cpub[:])
	if err != nil {
		return err
	}

	//Initialize a SecureReader and SecureWriter on top of the
	//connection and echo everything back to the client
	secread := NewSecureReader(conn, spriv, &cpub)
	secwrite := NewSecureWriter(conn, spriv, &cpub)
	buf := make([]byte, 32768)
	var n int
	for {
		n, err = secread.Read(buf)
		if err != nil {
			return err
		}
		_, err = secwrite.Write(buf[:n])
		if err != nil {
			return err
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
	if err != nil && err != io.EOF {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", buf[:n])
}
