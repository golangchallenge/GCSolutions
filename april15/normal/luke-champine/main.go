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

	"golang.org/x/crypto/nacl/box"
)

// A SecureReader reads and decrypts encrypted messages.
type SecureReader struct {
	io.Reader
	priv, pub *[32]byte
}

// Read implements the io.Reader interface.
func (sr *SecureReader) Read(p []byte) (int, error) {
	// read the nonce
	nonce := new([24]byte)
	n, err := io.ReadFull(sr.Reader, nonce[:])
	if err != nil {
		return n, err
	}
	// read the encrypted message
	// NOTE: NaCl introduces overhead. For this reason, we must read into a
	// buffer larger than p.
	msg := make([]byte, len(p)+box.Overhead)
	n, err = sr.Reader.Read(msg)
	if err != nil {
		return n, fmt.Errorf("failed to read message bytes: %v", err)
	}
	// decrypt the message
	// NOTE: box.Open appends the decrypted message to its first argument, so
	// here we reslice p to conserve memory.
	dec, ok := box.Open(p[:0], msg[:n], nonce, sr.pub, sr.priv)
	if !ok {
		return 0, errors.New("failed to decrypt message")
	}
	return len(dec), nil
}

// NewSecureReader instantiates a new SecureReader.
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	return &SecureReader{r, priv, pub}
}

// A SecureWriter writes encrypted messages. Messages are prefixed with a
// nonce, newly generated for each Write call.
type SecureWriter struct {
	io.Writer
	priv, pub *[32]byte
}

// Write implements the io.Writer interface.
func (sw *SecureWriter) Write(p []byte) (int, error) {
	// generate nonce
	nonce := new([24]byte)
	_, err := io.ReadFull(rand.Reader, nonce[:])
	if err != nil {
		return 0, err
	}
	// encrypt the message
	// NOTE: box.Seal appends the encrypted message to its first argument, so
	// for convenience we append to the nonce.
	enc := box.Seal(nonce[:], p, nonce, sw.pub, sw.priv)
	// write the encrypted message
	_, err = sw.Writer.Write(enc)
	return len(p), err
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	return &SecureWriter{w, priv, pub}
}

// Dial generates a private/public key pair, connects to the server, perform
// the handshake and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	// generate keypair
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	// connect to addr
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	// handshake: read the server's public key
	serverPub := new([32]byte)
	if _, err = io.ReadFull(conn, serverPub[:]); err != nil {
		return nil, err
	}
	// handshake: write our public key
	if _, err = conn.Write(pub[:]); err != nil {
		return nil, err
	}

	// Here we return an object that will satisfy the io.ReadWriteCloser
	// interface. Each type is embedded, which obviates the need for manually
	// defining Read/Write/Close methods. The Reader and Writer and simply
	// "Secure-" wrappings around conn.
	return struct {
		io.Reader
		io.Writer
		io.Closer
	}{
		NewSecureReader(conn, priv, serverPub),
		NewSecureWriter(conn, priv, serverPub),
		conn,
	}, nil
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	// generate keypair
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}
	// serve loop
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go func(conn net.Conn) {
			defer conn.Close()
			// handshake: write our public key
			if _, err := conn.Write(pub[:]); err != nil {
				log.Println("handshake failed:", err)
				return
			}
			// handshake: read the client's public key
			clientPub := new([32]byte)
			if _, err := io.ReadFull(conn, clientPub[:]); err != nil {
				log.Println("handshake failed:", err)
				return
			}
			// echo
			n, err := io.Copy(
				NewSecureWriter(conn, priv, clientPub),
				NewSecureReader(conn, priv, clientPub),
			)
			if err != nil {
				log.Println("echo failed:", err, n)
			}
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
