/*
Go Challenge 2
http://golang-challenge.com/go-challenge2/

In order to prevent our competitor from spying on our network,
we are going to write a small system that leverages NaCl to
establish secure communication. NaCl is a crypto system that
uses a public key for encryption and a private key for decryption.
*/
package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
)

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	myPub, myPriv, err := GenerateKey()
	if err != nil {
		return nil, err
	}

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	peerPub, err := clientHandshake(conn, myPub)
	if err != nil {
		return nil, err
	}

	srwc := NewSecureReadWriteCloser(conn, myPriv, peerPub)

	return srwc, nil
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	myPub, myPriv, err := GenerateKey()
	if err != nil {
		return err
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			continue
		}
		peerPub, err := serverHandshake(conn, myPub)
		if err != nil {
			continue
		}
		c := newConn(conn, myPriv, peerPub)
		c.serve()
	}

	return err
}

// clientHandshake writes it's pub key and reads the servers pub key
// from a net.Conn
func clientHandshake(conn net.Conn, pub *[32]byte) (*[32]byte, error) {
	// write my pub key to connection
	_, err := fmt.Fprintf(conn, string(pub[:32]))
	if err != nil {
		return nil, err
	}

	// read server pub key from connection
	buf := make([]byte, 32)
	_, err = conn.Read(buf)
	if err != nil {
		return nil, err
	}

	var peerPub [32]byte
	copy(peerPub[:], buf[0:32])
	return &peerPub, nil
}

// serverHandshake reads the client's pub key and writes it's pub key
// from a net.Conn
func serverHandshake(conn net.Conn, pub *[32]byte) (*[32]byte, error) {
	// read client pub key from connection
	buf := make([]byte, 32)
	_, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}

	// write my pub key to connection
	_, err = fmt.Fprintf(conn, string(pub[:32]))
	if err != nil {
		return nil, err
	}

	var peerPub [32]byte
	copy(peerPub[:], buf[0:32])
	return &peerPub, nil
}

// Create new connection from rwc.
func newConn(rwc net.Conn, priv, pub *[32]byte) *conn {
	c := new(conn)
	c.remoteAddr = rwc.RemoteAddr().String()
	c.rwc = rwc
	c.srwc = NewSecureReadWriteCloser(rwc, priv, pub)
	return c
}

// A conn represents the server side of an secure connection.
type conn struct {
	remoteAddr string
	rwc        net.Conn
	srwc       io.ReadWriteCloser
}

// Serve a new connection.
func (c *conn) serve() {
	// read from the client
	buf := make([]byte, maxMessageSize)
	n, err := c.srwc.Read(buf)
	if err != nil {
		fmt.Println("read error:", err)
	}
	buf = buf[:n]

	// write to the client connection
	_, err = fmt.Fprintf(c.srwc, string(buf))
	if err != nil {
		fmt.Println("write error:", err)
	}
}

func usage() {
	log.Fatalf("Usage: %s <port> <message>", os.Args[0])
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
	var message []byte
	var err error

	if len(os.Args) == 3 {
		message = []byte(os.Args[2])
	} else if len(os.Args) == 2 {
		message, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal(err)
		} else if len(message) == 0 {
			usage()
		}
	} else {
		usage()
	}

	conn, err := Dial("localhost:" + os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	if _, err = conn.Write(message); err != nil {
		log.Fatal(err)
	}

	buf := make([]byte, len(message))
	n, err := conn.Read(buf)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", buf[:n])
}
