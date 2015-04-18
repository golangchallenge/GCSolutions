package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"golang.org/x/crypto/nacl/box"
)

// If you're looking for NewSecureReader and NewSecureWriter, they're in secure.go (it's easier to read from top to bottom)

// PerformHandshake performs a key exchange with the underlying stream and returns a secure version of it
// rwc is the underlying ReadWriteCloser we want to do the handshake on
func PerformHandshake(rwc io.ReadWriteCloser) (*SecureReadWriteCloser, error) {
	ourPublicKey, ourPrivateKey, err := box.GenerateKey(new(CryptoRandomReader))
	var theirPublicKey [32]byte

	_, err = rwc.Write(ourPublicKey[:])
	if err != nil {
		return nil, err
	}

	_, err = io.ReadFull(rwc, theirPublicKey[:])
	if err != nil {
		return nil, err
	}
	return NewSecureReadWriteCloser(rwc, ourPrivateKey, &theirPublicKey), nil
}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	return PerformHandshake(conn)
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go func(conn net.Conn) {
			defer conn.Close()

			sconn, err := PerformHandshake(conn)
			if err != nil {
				fmt.Println(err)
				return
			}

			msg, err := sconn.ReadMsg()
			if err != nil {
				log.Println(err)
				return
			}

			_, err = sconn.Write(msg.Data)
			if err != nil {
				log.Println(err)
				return
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
			return
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
	defer conn.Close()

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
