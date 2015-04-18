package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

func generateKey() (*[32]byte, error) {
	rbytes := make([]byte, 32)
	_, err := rand.Read(rbytes)
	key := &[32]byte{}
	copy(key[:], rbytes)
	return key, err
}

// Dial generates a private/public key pair,
// connects to the server, performs the handshake,
// and returns a reader/writer/closer.
func Dial(addr string) (rwc io.ReadWriteCloser, e error) {
	conn, e := net.Dial("tcp", addr)
	if e != nil {
		return
	}

	priv, e := generateKey()
	if e != nil {
		return
	}
	_, e = conn.Write(priv[:])
	if e != nil {
		return
	}

	pub, e := generateKey()
	if e != nil {
		return
	}
	_, e = conn.Write(pub[:])
	if e != nil {
		return
	}

	rwc = struct {
		io.Reader
		io.Writer
		io.Closer
	}{
		&SecureReader{conn, priv, pub},
		&SecureWriter{conn, priv, pub},
		conn,
	}
	return
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) (e error) {
	conn, e := l.Accept()
	if e != nil {
		return
	}

	defer conn.Close()

	privPub := make([]byte, 64)
	read, e := conn.Read(privPub)
	if e != nil || read != 64 {
		fmt.Fprintf(conn, "Handshake failure")
		return
	}

	priv := &[32]byte{}
	copy(priv[:], privPub[:32])

	pub := &[32]byte{}
	copy(pub[:], privPub[32:])

	reader := SecureReader{conn, priv, pub}

	// Read up to a 32KB message(decrypted)
	req := make([]byte, 32768)
	bytecount, e := reader.Read(req)
	if e != nil {
		return
	}

	writer := SecureWriter{conn, priv, pub}
	_, e = writer.Write(req[0:bytecount])
	return
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
