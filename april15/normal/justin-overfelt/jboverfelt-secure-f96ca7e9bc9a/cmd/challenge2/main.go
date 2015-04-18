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

	"bitbucket.org/jboverfelt/secure"
)

type secureConn struct {
	io.Reader
	io.Writer
	io.Closer
}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func dial(addr string) (io.ReadWriteCloser, error) {
	pub, priv, err := box.GenerateKey(rand.Reader)

	if err != nil {
		return nil, err
	}

	conn, err := net.Dial("tcp", addr)

	if err != nil {
		return nil, err
	}

	// first thing we do is send our public key
	_, err = conn.Write(pub[:])

	if err != nil {
		return nil, err
	}

	// wait for the server's public key
	peerPubSlice := make([]byte, secure.KeySize)
	n, err := conn.Read(peerPubSlice)

	if err != nil {
		return nil, err
	}

	peerPubSlice = peerPubSlice[:n]
	var peerPub [secure.KeySize]byte
	copy(peerPub[:], peerPubSlice)

	secCon := secureConn{
		secure.NewReader(conn, priv, &peerPub),
		secure.NewWriter(conn, priv, &peerPub),
		conn,
	}

	return secCon, nil
}

// Serve starts a secure echo server on the given listener.
func serve(l net.Listener) error {
	pub, priv, err := box.GenerateKey(rand.Reader)

	if err != nil {
		return err
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		go handleConnection(conn, pub, priv)
	}
}

func handleConnection(c net.Conn, pub, priv *[32]byte) {
	defer c.Close()
	// first wait for the client's public key
	peerPubSlice := make([]byte, secure.KeySize)
	n, err := c.Read(peerPubSlice)

	if err != nil {
		log.Println(err)
		return
	}

	peerPubSlice = peerPubSlice[:n]
	var peerPub [secure.KeySize]byte
	copy(peerPub[:], peerPubSlice)

	// then, send our public key
	_, err = c.Write(pub[:])

	if err != nil {
		log.Println(err)
		return
	}

	// now session is "secure"
	sr := secure.NewReader(c, priv, &peerPub)
	sw := secure.NewWriter(c, priv, &peerPub)

	// echo
	_, err = io.Copy(sw, sr)

	if err != nil {
		log.Println(err)
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
		log.Fatal(serve(l))
	}

	// Client mode
	if len(os.Args) != 3 {
		log.Fatalf("Usage: %s <port> <message>", os.Args[0])
	}
	conn, err := dial("localhost:" + os.Args[1])
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
