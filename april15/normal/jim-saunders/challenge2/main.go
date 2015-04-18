package main

import (
	"crypto/rand"
	"encoding/binary"
	"flag"
	"fmt"
	"golang.org/x/crypto/nacl/box"
	"io"
	"log"
	"net"
	"os"
)

// NewSecureReadWriteCloser instantiates an io.ReadWriteCloser with a SecureReader and SecureWriter
func NewSecureReadWriteCloser(conn io.ReadWriteCloser, priv, pub *[32]byte) io.ReadWriteCloser {
	return struct {
		io.Reader
		io.Writer
		io.Closer
	}{
		SecureReader{conn, priv, pub},
		SecureWriter{conn, priv, pub},
		conn,
	}
}

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	return SecureReader{r, priv, pub}
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	return SecureWriter{w, priv, pub}
}

// Generate public and private keys send the generated
// public key to the provided connection and read a
// public key from the provided connection.
func exchangeKeys(conn io.ReadWriter) (*[32]byte, *[32]byte, error) {
	send, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	if err := binary.Write(conn, binary.LittleEndian, send); err != nil {
		return nil, nil, err
	}

	var recv [32]byte
	if err := binary.Read(conn, binary.LittleEndian, &recv); err != nil {
		return nil, nil, err
	}

	return priv, &recv, nil
}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	priv, pub, err := exchangeKeys(conn)
	if err != nil {
		return nil, err
	}

	return NewSecureReadWriteCloser(conn, priv, pub), nil
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	for {

		conn, err := l.Accept()
		if err != nil {
			return err
		}

		priv, pub, err := exchangeKeys(conn)
		if err != nil {
			return err
		}

		io.Copy(
			NewSecureWriter(conn, priv, pub),
			NewSecureReader(conn, priv, pub),
		)
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
