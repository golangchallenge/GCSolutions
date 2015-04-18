package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

const (
	maxPayload = 32768
)

type SecureReadWriter struct {
	r    io.Reader
	w    io.Writer
	conn net.Conn
}

func (srw *SecureReadWriter) Close() error {
	return srw.conn.Close()
}

func (srw *SecureReadWriter) Read(p []byte) (n int, err error) {
	return srw.r.Read(p)
}

func (srw *SecureReadWriter) Write(p []byte) (n int, err error) {
	return srw.w.Write(p)
}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Printf("error connecting at %v:%v\n", addr, err)
		return nil, nil
	}

	pub, priv, err := GenerateKey()

	if err != nil {
		log.Printf("error generating key pair:%v\n", err)
		return nil, err
	}

	srvKey := new([32]byte)
	err = keyExchange(conn, pub, srvKey)
	if err != nil {
		log.Printf("handshake error:%v\n", err)
		return nil, err
	}

	srw := &SecureReadWriter{NewSecureReader(conn, priv, srvKey), NewSecureWriter(conn, priv, srvKey), conn}

	return srw, nil
}

func keyExchange(conn net.Conn, pub, key *[32]byte) error {
	//key exchange
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := sendKey(conn, pub)
		if err != nil {
			log.Printf("%v", err)
		}
	}()
	err := readKey(conn, key)
	wg.Wait()
	//key exchange end
	return err
}

func sendKey(w io.Writer, key *[32]byte) error {
	n, err := w.Write(key[:])
	if err != nil {
		log.Printf("error sending pubkey[%v] to server:%v\n", key, err)
		return err
	}
	if n != 32 {
		err = errors.New(fmt.Sprintf("expected 32 but sent %v", n))
		log.Printf("%v\n", err)
		return err
	}
	return nil
}

func readKey(r io.Reader, key *[32]byte) error {
	n, err := r.Read(key[:])
	if err != nil {
		log.Printf("error reading pubkey to server:%v\n", err)
		return err
	}

	if n != 32 {
		err = errors.New(fmt.Sprintf("expected 32 but recieved %v", n))
		return err
	}
	return nil
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	pub, priv, err := GenerateKey()

	if err != nil {
		log.Printf("error server generating key pair:%v\n", err)
		return err
	}

	conn, err := l.Accept()
	if err != nil {
		log.Printf("error accepting conn:%v\n", err)
		return err
	}

	cliKey := new([32]byte)
	err = keyExchange(conn, pub, cliKey)
	if err != nil {
		log.Printf("handshake error:%v\n", err)
		return err
	}

	sr := NewSecureReader(conn, priv, cliKey)

	//just to sync clientID
	NewSecureWriter(conn, priv, cliKey)

	payload := new([maxPayload]byte)

	n, err := sr.Read(payload[:])

	//log.Printf("%v\n", string(payload[0:n]))

	conn.Write(payload[0:n])

	return nil
}
