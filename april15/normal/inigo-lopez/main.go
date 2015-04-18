package main

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"golang.org/x/crypto/nacl/box"
)

const (
	maxSize = 24 + 2 + 32*1024 + box.Overhead // nonce + size bytes + max message size + overhead
)

// A SecureReader is an io.Reader that uses NaCl to decrypt messages.
type SecureReader struct {
	r    io.Reader // Underlying io.Reader.
	priv *[32]byte // Private key.
	ppub *[32]byte // Peer's public key.
}

// Read decrypts a message from the underlying data stream, and stores
// the contents into out, returning the number of bytes read, and any
// error occurred while reading the data. The encrypted data contains
// the nonce (24 bytes), box size (2 bytes), and box (remaining bytes).
func (r *SecureReader) Read(out []byte) (n int, err error) {
	// "Sticky error" read function.
	read := func(buf []byte) {
		if err != nil {
			return
		}
		_, err = io.ReadFull(r.r, buf)
	}

	var (
		nonce [24]byte
		size  = make([]byte, 2)
	)

	read(nonce[:])
	read(size)
	boxed := make([]byte, binary.LittleEndian.Uint16(size)) // Allocate the box bytes.
	read(boxed)

	if err != nil {
		return
	}

	message, ok := box.Open(out[:0], boxed, &nonce, r.ppub, r.priv)
	if !ok {
		err = errors.New("cannot open box")
	}
	n = len(message)
	return
}

// A SecureWriter is an io.Writer that uses NaCl to encrypt messages.
type SecureWriter struct {
	w    io.Writer // Underlying io.Writer.
	priv *[32]byte // Private key.
	ppub *[32]byte // Peer's public key.
}

// Write encrypts p and writes the contents to the underlying data stream,
// returning the number of bytes written from p (either 0 or len(p)), and
// any error occurred while writing the data. The encrypted data contains
// the nonce (24 bytes), box size (2 bytes), and box (len(p)+box.Overhead).
func (w *SecureWriter) Write(p []byte) (n int, err error) {
	// Generate a random nonce.
	var nonce [24]byte
	_, err = rand.Read(nonce[:])

	// Encrypt the data, getting a freshly allocated []byte back.
	box := box.Seal(nil, p, &nonce, w.ppub, w.priv)

	// A uint16(2 bytes) is enough to handle a maximum size of 32*1024 bytes messages.
	size := make([]byte, 2)
	binary.LittleEndian.PutUint16(size, uint16(len(box)))

	// "Sticky error" write function.
	write := func(buf []byte) {
		if err != nil {
			return
		}
		_, err = w.w.Write(buf)
	}

	// Write the data: nonce(24 bytes), size(2 bytes), box ('size' bytes).
	write(nonce[:])
	write(size)
	write(box)
	if err == nil {
		n = len(p)
	}
	return
}

// NewSecureConn returns a secure io.ReadWriteCloser using conn for I/O.
func NewSecureConn(conn net.Conn, priv, pub *[32]byte) io.ReadWriteCloser {
	return struct {
		io.Reader
		io.Writer
		io.Closer
	}{
		NewSecureReader(conn, priv, pub),
		NewSecureWriter(conn, priv, pub),
		conn,
	}
}

// NewSecureReader instantiates a new SecureReader.
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	return &SecureReader{r, priv, pub}
}

// NewSecureWriter instantiates a new SecureWriter.
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	return &SecureWriter{w, priv, pub}
}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	// Generate keys.
	pub, priv, err := box.GenerateKey(rand.Reader)

	// Read peer's public key.
	var ppub [32]byte
	if _, err = conn.Read(ppub[:]); err != nil {
		return nil, err
	}
	// Send public key.
	if _, err = conn.Write(pub[:]); err != nil {
		return nil, err
	}

	return NewSecureConn(conn, priv, &ppub), nil
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	// Generate a private/public key pair
	pub, priv, _ := box.GenerateKey(rand.Reader)

	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		go func(c net.Conn) {
			defer c.Close()

			// Send public key.
			if _, err := c.Write(pub[:]); err != nil {
				log.Println(err)
				return
			}
			// Read peer's public key.
			var ppub [32]byte
			if _, err := c.Read(ppub[:]); err != nil {
				log.Println(err)
				return
			}

			// Using the secured io.ReadWriteCloser, decrypt and encrypt the message
			sc := NewSecureConn(c, priv, &ppub)

			buf := make([]byte, maxSize)
			n, err := sc.Read(buf)
			if err != nil {
				log.Println(err)
				return
			}
			buf = buf[:n]

			if _, err = sc.Write(buf); err != nil {
				log.Println(err)
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
