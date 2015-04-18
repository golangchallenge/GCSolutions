package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	"golang.org/x/crypto/nacl/box"
)

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	// make local keys
	conn, err := net.Dial("tcp", addr)

	if err != nil {
		return nil, err
	}

	publicKey, privateKey, err := box.GenerateKey(rand.Reader)

	if err != nil {
		return nil, err
	}

	var peersPublicKey [32]byte

	// do kex
	_, err = conn.Write(publicKey[:])
	if err != nil {
		return nil, err
	}
	_, err = io.ReadFull(conn, peersPublicKey[:])
	if err != nil {
		return nil, err
	}

	ret := NewSecureReadWriteCloser(conn, privateKey, &peersPublicKey)

	return ret, nil
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go ServeOne(conn)
	}
}

// ServeOne serves one network connection
// It generats local keys, then performs the key exchange (simply reads the
// remote ends public key and writes its own public key) then reads the secure
// message and sends it back
func ServeOne(conn net.Conn) {

	defer conn.Close()

	// make local keys
	publicKey, privateKey, _ := box.GenerateKey(rand.Reader)
	var peersPublicKey [32]byte

	// do kex
	_, err := conn.Write(publicKey[:])
	if err != nil {
		return
	}

	conn.SetReadDeadline(time.Now().Add(time.Second))

	_, err = io.ReadFull(conn, peersPublicKey[:])
	if err != nil {
		return
	}

	secureR := NewSecureReader(conn, privateKey, &peersPublicKey)
	secureW := NewSecureWriter(conn, privateKey, &peersPublicKey)

	// read msg
	buf := make([]byte, maxMessageLen)
	//for {
	n, err := secureR.Read(buf)
	if err != nil {
		return
	}

	// write msg back
	secureW.Write(buf[:n])
	//}
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
