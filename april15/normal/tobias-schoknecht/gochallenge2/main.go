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

// SecureReader is a wrapper type around any io.Reader.
// It automatically decrypts received messages using the shared key.
type SecureReader struct {
	r      io.Reader
	shared *[32]byte
}

func (sr SecureReader) Read(decryptedMsg []byte) (n int, err error) {
	nonce := [24]byte{}

	_, err = sr.r.Read(nonce[:])
	if err != nil {
		return 0, err
	}

	msg := [32768]byte{}
	l, err := sr.r.Read(msg[:])
	if err != nil {
		return 0, err
	}
	cryptedMsg := msg[:l]

	// The length of the resulting message is unknown.
	// Passing an empty out will result in everything being passed to overlap
	overlap, b := box.OpenAfterPrecomputation([]byte{}, cryptedMsg, &nonce, sr.shared)
	if !b {
		return 0, errors.New("Decrypt failed")
	}
	copy(decryptedMsg, overlap)

	return len(overlap), nil
}

// SecureWriter is a wrapper type around any io.Writer.
// It automatically encrypts received messages using the shared key.
type SecureWriter struct {
	w      io.Writer
	shared *[32]byte
}

func (sw SecureWriter) Write(unencryptedMsg []byte) (n int, err error) {
	nonce := [24]byte{}
	_, err = rand.Read(nonce[:])
	if err != nil {
		return 0, err
	}
	cryptedMsg := box.SealAfterPrecomputation([]byte{}, unencryptedMsg, &nonce, sw.shared)

	msg := []byte{}
	msg = append(msg, nonce[:]...)
	msg = append(msg, cryptedMsg...)

	return sw.w.Write(msg)
}

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	shared := &[32]byte{}
	box.Precompute(shared, pub, priv)
	return SecureReader{r: r, shared: shared}
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	shared := &[32]byte{}
	box.Precompute(shared, pub, priv)
	return SecureWriter{w: w, shared: shared}
}

type SecureConnection struct {
	io.Reader
	io.Writer
	io.Closer
}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	pubClient, privClient, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	// Send the public key to the server
	if _, err = conn.Write(pubClient[:]); err != nil {
		return nil, err
	}
	// Reads the public key from the server
	pubServer := [32]byte{}
	if _, err = conn.Read(pubServer[:]); err != nil {
		return nil, err
	}

	return SecureConnection{
		Reader: NewSecureReader(conn, privClient, &pubServer),
		Writer: NewSecureWriter(conn, privClient, &pubServer),
		Closer: conn,
	}, nil
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	for {
		// Wait for a connection.
		conn, err := l.Accept()
		if err != nil {
			return nil
		}

		// Handle the connection in a new goroutine.
		// The loop then returns to accepting, so that
		// multiple connections may be served concurrently.
		go func(c net.Conn) {
			pubServer, privServer, err := box.GenerateKey(rand.Reader)
			if err != nil {
				fmt.Println(err)
				return
			}

			// Read the public key from the client
			pubClient := [32]byte{}
			if _, err := c.Read(pubClient[:]); err != nil {
				fmt.Println(err)
				return
			}

			// Send the server public key to the client
			if _, err := c.Write(pubServer[:]); err != nil {
				fmt.Println(err)
				return
			}

			reader := NewSecureReader(c, privServer, &pubClient)
			writer := NewSecureWriter(c, privServer, &pubClient)

			// The general assumption is that messages are never bigger
			// than 32KB
			buf := make([]byte, 32768)
			reqLen, err := reader.Read(buf)
			if err != nil {
				fmt.Println("Error reading:", err.Error())
			}
			if _, err = writer.Write(buf[:reqLen]); err != nil {
				fmt.Println("Failed to write to client")
			}

			c.Close()
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
