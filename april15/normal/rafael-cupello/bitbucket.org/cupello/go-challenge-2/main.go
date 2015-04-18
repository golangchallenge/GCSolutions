package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"bitbucket.org/cupello/go-challenge-2/secureio"
)

const (
	// max expected message size : 32KB
	maxExpectedMessageSize = 1024 * 32
)

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, peerPub *[32]byte) io.Reader {
	return secureio.NewSecureReader(r, priv, peerPub)
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, peerPub *[32]byte) io.Writer {
	return secureio.NewSecureWriter(w, priv, peerPub)
}

// NewSecureReadWriter instantiates a new SecureReadWriter
func NewSecureReadWriter(rw io.ReadWriter, priv, peerPub *[32]byte) io.ReadWriter {
	return secureio.NewSecureReadWriter(rw, priv, peerPub)
}

// NewSecureReadWriteCloser instantiates a new SecureReadWriteCloser
func NewSecureReadWriteCloser(rwc io.ReadWriteCloser, priv, peerPub *[32]byte) io.ReadWriteCloser {
	return secureio.NewSecureReadWriteCloser(rwc, priv, peerPub)
}

// exchangeKeys performs the exchange of the keys
// by writing the public key into the data stream first and
// then reading the shared key from the data stream.
func exchangeKeys(pub, peerPub *[32]byte, rw io.ReadWriter) error {
	if _, err := rw.Write(pub[:]); err != nil {
		return err
	}

	if _, err := rw.Read(peerPub[:]); err != nil {
		return err
	}

	return nil
}

// handleServerConnection hands a server connection.
func handleServerConnection(pub, priv *[32]byte, conn net.Conn) {
	defer conn.Close()

	// Exchange public key with the client
	peerPub := new([32]byte)
	if err := exchangeKeys(pub, peerPub, conn); err != nil {
		log.Fatal(err)
		return
	}

	// Create a new SecureReadWriteCloser
	rw := NewSecureReadWriter(conn, priv, peerPub)

	// Read the message from the data stream into a buffer
	buf := make([]byte, maxExpectedMessageSize)
	ln, err := rw.Read(buf)
	if err != nil {
		log.Fatal(err)
		return
	}

	// Write the message from the buffer to the data stream
	if _, err := rw.Write(buf[:ln]); err != nil {
		log.Fatal(err)
		return
	}
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {

	pub, priv, err := secureio.GenerateKeyPair()
	if err != nil {
		return err
	}

	// Start the server
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go handleServerConnection(pub, priv, conn)
	}
}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {

	pub, priv, err := secureio.GenerateKeyPair()
	if err != nil {
		return nil, err
	}

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	peerPub := new([32]byte)
	if err := exchangeKeys(pub, peerPub, conn); err != nil {
		return nil, err
	}

	return NewSecureReadWriteCloser(conn, priv, peerPub), nil
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
