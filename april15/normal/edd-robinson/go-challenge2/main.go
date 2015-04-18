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

const (
	nsize      = 24
	ksize      = 32
	maxDecSize = 32<<10 - 1 // Messages are less than 32KB.
	minBoxSize = 1 + box.Overhead
	maxBoxSize = maxDecSize + box.Overhead
)

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	var (
		myPub, myPriv *[ksize]byte
		peerPub       = new([ksize]byte)
		err           error
	)

	// Generate our own keypair.
	if myPub, myPriv, err = box.GenerateKey(rand.Reader); err != nil {
		return nil, err
	}

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	// Write the our public key to the connection.
	if _, err = conn.Write(myPub[:]); err != nil {
		return nil, err
	}

	// Read the server's public key from the connection; ReadFull will
	// return an error if we don't get exactly ksize bytes.
	if _, err = io.ReadFull(conn, peerPub[:]); err != nil {
		return nil, err
	}

	return struct {
		io.Reader
		io.Writer
		io.Closer
	}{
		NewSecureReader(conn, myPriv, peerPub),
		NewSecureWriter(conn, myPriv, peerPub),
		conn,
	}, nil
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	var (
		r             io.Reader
		w             io.Writer
		myPub, myPriv *[ksize]byte
		peerPub       = new([ksize]byte)
	)

	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		// Generate our own keypair for the connection.
		if myPub, myPriv, err = box.GenerateKey(rand.Reader); err != nil {
			return err
		}

		// Write our public key to the connection.
		if _, err = conn.Write(myPub[:]); err != nil {
			return err
		}

		// Read the client's public key from the connection; ReadFull
		// will return an error if we don't get exactly ksize bytes.
		if _, err = io.ReadFull(conn, peerPub[:]); err != nil {
			return err
		}

		go func() {
			defer conn.Close()
			r = NewSecureReader(conn, myPriv, peerPub)
			w = NewSecureWriter(conn, myPriv, peerPub)
			if _, err := io.Copy(w, r); err != nil {
				log.Println(err)
			}
		}()
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
