package main

import (
	"crypto/rand"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"golang.org/x/crypto/nacl/box"
)

// writerFunc allows any function with the proper signature to implement io.Writer.
type writerFunc func([]byte) (int, error)

func (w writerFunc) Write(msg []byte) (int, error) {
	return w(msg)
}

// readerFunc allows any function with the proper signature to implement io.Reader.
type readerFunc func([]byte) (int, error)

func (r readerFunc) Read(p []byte) (int, error) {
	return r(p)
}

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	return readerFunc(func(p []byte) (int, error) {
		var size int32
		//first read size of payload
		binary.Read(r, binary.BigEndian, &size)
		if int32(len(p)) < size {
			return 0, fmt.Errorf("Supplied buffer of size %d not large enough for message of size %d.", len(p), size)
		}
		//make buffer big enough for nonce + payload + overhead
		buf := make([]byte, size+24+box.Overhead)
		n, err := io.ReadFull(r, buf)
		if err != nil {
			return n, err
		}
		nonce := [24]byte{}
		copy(nonce[:], buf[:24])
		msg := buf[24:n]
		output, ok := box.Open(nil, msg, &nonce, pub, priv)
		if !ok {
			return 0, fmt.Errorf("Error decrypting message")
		}
		n = copy(p, output)
		return n, err
	})
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	return writerFunc(func(msg []byte) (int, error) {
		nonce := [24]byte{}
		_, err := rand.Read(nonce[:])
		if err != nil {
			return 0, err
		}
		out := box.Seal(nil, msg, &nonce, pub, priv)
		if err != nil {
			return 0, err
		}
		// write [message length | nonce | encrypted message]
		err = binary.Write(w, binary.BigEndian, int32(len(msg)))
		if err != nil {
			fmt.Println(err)
			return 0, err
		}
		n, err := w.Write(append(nonce[:], out...))
		return n + 4, err
	})
}

// Allows compositon of three seperate items to form a single ReadWriteCloser.
type readWriteCloser struct {
	io.Reader
	io.Writer
	io.Closer
}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	pub, priv, err := generateKeys()
	if err != nil {
		return nil, err
	}
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	serverPub, err := exchangeKeys(pub, conn)
	if err != nil {
		return nil, err
	}
	reader := NewSecureReader(conn, priv, serverPub)
	writer := NewSecureWriter(conn, priv, serverPub)
	return readWriteCloser{reader, writer, conn}, nil
}

func exchangeKeys(pub *[32]byte, conn io.ReadWriter) (*[32]byte, error) {
	_, err := conn.Write(pub[:])
	if err != nil {
		return nil, err
	}

	otherPub := [32]byte{}
	_, err = conn.Read(otherPub[:])
	if err != nil {
		return nil, err
	}
	return &otherPub, nil
}

func generateKeys() (pub, priv *[32]byte, err error) {
	return box.GenerateKey(rand.Reader)
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	pub, priv, err := generateKeys()
	if err != nil {
		return err
	}
	for {
		c, err := l.Accept()
		if err != nil {
			return err
		}
		go serveConnection(c, pub, priv)
	}
}

func serveConnection(conn net.Conn, pub, priv *[32]byte) (err error) {
	defer conn.Close()
	defer func() {
		if err != nil {
			log.Println("Error serving connection:", err)
		}
	}()
	clientPub, err := exchangeKeys(pub, conn)
	// read a message from the client
	reader := NewSecureReader(conn, priv, clientPub)
	buf := make([]byte, 1024*32)
	n, err := reader.Read(buf)
	if err != nil {
		return err
	}
	fmt.Println(string(buf[:n]))

	writer := NewSecureWriter(conn, priv, clientPub)
	n, err = writer.Write(buf[:n])
	if err != nil {
		return err
	}
	return nil
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
		log.Printf("listening on port %d.\n", *port)
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
