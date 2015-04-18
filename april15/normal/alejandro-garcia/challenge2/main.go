// Package main provides the server, client dial and NewSecure io's functions
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

	"challenge2/helper"
	"challenge2/nio"

	"golang.org/x/crypto/nacl/box"
)

// NewSecureReader instantiates a secure io.Reader
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	shared := &[32]byte{}
	// Let's get a bit faster
	box.Precompute(shared, pub, priv)
	return nio.NaCLReader{Reader: r, Shared: shared}
}

// NewSecureWriter instantiates a secure io.Writer
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	shared := &[32]byte{}
	// Let's get a bit faster
	box.Precompute(shared, pub, priv)
	return nio.NaCLWriter{Writer: w, Shared: shared}
}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	// public and private keys
	pub, priv, err := box.GenerateKey(rand.Reader)
	if *priv == *pub {
		return nil, errors.New("keys are equal")
	}
	if err != nil {
		return nil, err
	}
	// connect to server
	conn, err := net.Dial("tcp", addr)
	// send public key to server unencrypted
	_, err = conn.Write(pub[:])
	if err != nil {
		return nil, err
	}
	// read server side public key unencrypted
	spubk := make([]byte, 1024)
	n, err := conn.Read(spubk)
	if err != nil {
		return nil, err
	}
	// transform server public key into [32]byte
	spub := &[32]byte{}
	copy(spub[:], spubk[:n])

	nio.Nonce = helper.Nonce{}

	rd := NewSecureReader(conn, priv, spub)
	wr := NewSecureWriter(conn, priv, spub)

	// init read,writer,closer
	rwc := nio.NaCLReaderWriterCloser{Rd: rd, Wr: wr, Cl: conn}
	return rwc, nil
}

// handleConnection is the server side function
// handles the public keys exchange as handshake and message serving
func handleConnection(conn net.Conn) error {
	defer conn.Close()
	// Handshake
	// Generate server side keys
	pub, pri, err := box.GenerateKey(rand.Reader)
	if *pri == *pub || err != nil {
		return err
	}

	// cpubk is the client public key
	cpubk := make([]byte, 2048)
	n, err := conn.Read(cpubk)
	if err != nil {
		return err
	}
	// client public key
	cpub := &[32]byte{}
	copy(cpub[:], cpubk[:n])

	// send server public key to client
	_, err = conn.Write(pub[:])
	if err != nil {
		return err
	}

	shared := &[32]byte{}
	// Let's get a bit faster
	box.Precompute(shared, cpub, pri)

	// reading message
	msg := make([]byte, 2048)
	n, err = conn.Read(msg)
	if err != nil {
		fmt.Errorf("could not read message %s", err)
		return err
	}
	// Generate a nonc, to keep it simple we synchronize
	// nonces between Clietn and Server
	nonc := &helper.Nonce{}
	nv := nonc.GenerateNonce(*shared)

	out := make([]byte, n)
	// decrypt message
	dm, ok := box.OpenAfterPrecomputation(out, msg[:n], &nv, shared)
	if !ok {
		m := fmt.Sprintf("could not decrypt message %v", msg[:n])
		return errors.New(m)
	}

	// Send back message encrypted
	em := box.SealAfterPrecomputation(nil, dm[n:], &nv, shared)
	_, err = conn.Write(em)
	if err != nil {
		return err
	}
	return nil
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go func() {
			err := handleConnection(conn)
			if err != nil {
				log.Printf("Handling %s\n", err)
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
	defer conn.Close()
	if _, err := conn.Write([]byte(os.Args[2])); err != nil {
		log.Fatalf("Writing problem %s\n", err.Error())
	}
	buf := make([]byte, 2048)
	n, err := conn.Read(buf)
	if err != nil {
		log.Fatalf("Reading problem %s\n", err.Error())
	}
	fmt.Printf("%s\n", buf[:n])
}
