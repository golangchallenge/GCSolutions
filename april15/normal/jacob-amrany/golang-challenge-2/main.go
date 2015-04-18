package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"./nacl"
	"golang.org/x/crypto/nacl/box"
)

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	return nacl.NewSecureReader(r, priv, pub)
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	return nacl.NewSecureWriter(w, priv, pub)
}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	_, err = conn.Write(pub[:])
	if err != nil {
		return nil, err
	}

	var peer [32]byte
	n, err := conn.Read(peer[:])
	if err != nil || n != 32 {
		return nil, err
	}

	rwc := nacl.NewRWC(conn, priv, &peer)

	return rwc, err
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}
	conn, err := l.Accept()
	if err != nil {
		return err
	}

	//Handshake Receive public key and private key
	var peer [32]byte
	_, err = conn.Read(peer[:])
	if err != nil {
		return err
	}

	_, err = conn.Write(pub[:])
	if err != nil {
		return err
	}

	rwc := nacl.NewRWC(conn, priv, &peer)

	err = echo(rwc)
	return err
}

func echo(rwc io.ReadWriteCloser) (err error) {
	/*
		_, err = io.Copy(rwc, rwc)
		if err != nil {
			return err
		}
		return err
	*/
	buf := make([]byte, 1048)
	n, err := rwc.Read(buf)
	if err != nil {
		fmt.Println("read completed", err)
		return err
	}
	buf = buf[:n]
	_, err = rwc.Write(buf)
	if err != nil {
		return err
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
