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

	// Exchange the keys and nonce.
	nonce := NewNonce()
	rpub := &[32]byte{}

	if _, err = conn.Write(pub[:]); err != nil {
		return nil, err
	}
	if _, err = conn.Write(nonce[:]); err != nil {
		return nil, err
	}
	if _, err = conn.Read(rpub[:]); err != nil {
		return nil, err
	}

	secureConn := NewSecureReadWriteCloser(conn, priv, rpub, nonce)
	return secureConn, nil
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) (err error) {
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		go func(c net.Conn) {
			defer c.Close()

			rpub := &[32]byte{}
			nonce := &[24]byte{}

			if _, err = conn.Write(pub[:]); err != nil {
				panic("Serve - public key exchange failed")
			}
			if _, err = conn.Read(rpub[:]); err != nil {
				panic("Serve - remote public key exchange failed")
			}
			if _, err = conn.Read(nonce[:]); err != nil {
				panic("Serve - nonce exchange failed")
			}

			secureConn := NewSecureReadWriteCloser(conn, priv, rpub, *nonce)
			io.Copy(secureConn, secureConn)
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
