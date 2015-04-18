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

var (
	// ErrDecryptFailed is returned by SecureReader.Read when it can't decrypt the received data.
	ErrDecryptFailed = errors.New("decrypt failed")
	// ErrMessageTooLong is returned by SecureWriter.Write when the size of the passed slice exceeds the maximum.
	ErrMessageTooLong = errors.New("message too long; maximum length is 65519 bytes")
)

// SecureReader reads and decrypts data from an io.Reader object.
type SecureReader struct {
	rd        io.Reader
	sharedKey [32]byte
	buf       []byte
	r, w      int // read and write positions in the buffer
}

// NewSecureReader instantiates a new SecureReader.
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	sr := SecureReader{
		rd:  r,
		buf: make([]byte, 64*1024),
	}
	box.Precompute(&sr.sharedKey, pub, priv) // precompute a shared key
	return &sr
}

// Read reads from the encrypted stream.
func (r *SecureReader) Read(p []byte) (n int, err error) {
	// If there are bytes in the buffer, copy them into p first.
	if r.r < r.w {
		n = copy(p, r.buf[r.r:r.w])
		r.r += n
		if n == len(p) {
			return
		}
		p = p[n:]
	}
	r.r, r.w = 0, 0 // reset the buffer

	// Read a nonce.
	nonce := new([24]byte)
	_, err = io.ReadFull(r.rd, nonce[:])
	if err != nil {
		return
	}

	// Read the encrypted data length.
	var length uint16
	err = binary.Read(r.rd, binary.LittleEndian, &length)
	if err != nil {
		return
	}

	// Read the encrypted data.
	buf := make([]byte, length)
	_, err = io.ReadFull(r.rd, buf)
	if err != nil {
		return
	}

	// Decrypt the data.
	opened, ok := box.OpenAfterPrecomputation(nil, buf, nonce, &r.sharedKey)
	if !ok {
		err = ErrDecryptFailed
		return
	}

	// Write the decrypted data to the provided slice.
	n += copy(p, opened)

	// If the size of p isn't enough, buffer the remaining bytes.
	if len(p) < len(opened) {
		r.w = copy(r.buf, opened[len(p):])
	}

	return
}

// SecureWriter encrypts data and writes it to the underlying io.Writer object.
type SecureWriter struct {
	w         io.Writer
	sharedKey [32]byte
}

// NewSecureWriter instantiates a new SecureWriter.
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	sw := SecureWriter{w: w}
	box.Precompute(&sw.sharedKey, pub, priv) // precompute a shared key
	return &sw
}

// Write writes an encrypted message with the contents of p to the underlying writer.
// The format of the message is as follows:
// - nonce (24 bytes)
// - length of the following encrypted data (2 bytes in the little-endian byte order)
// - encrypted data
func (w *SecureWriter) Write(p []byte) (n int, err error) {
	if len(p) > 0xffff-box.Overhead {
		return 0, ErrMessageTooLong
	}

	// Generate a random nonce.
	var nonce [24]byte
	_, err = io.ReadFull(rand.Reader, nonce[:])
	if err != nil {
		return
	}

	// Encrypt the data.
	sealed := box.SealAfterPrecomputation(nil, p, &nonce, &w.sharedKey)

	// Write the nonce.
	_, err = w.w.Write(nonce[:])
	if err != nil {
		return
	}

	// Write the length.
	err = binary.Write(w.w, binary.LittleEndian, uint16(len(sealed)))
	if err != nil {
		return
	}

	// Write the encrypted data.
	_, err = w.w.Write(sealed)
	if err != nil {
		return
	}

	return len(p), nil
}

// SecureReadWriter is a wrapper around net.Conn for "secure" communication.
type SecureReadWriter struct {
	conn net.Conn
	io.Reader
	io.Writer
}

// Close implements io.Closer.
func (rw *SecureReadWriter) Close() error {
	return rw.conn.Close()
}

// Dial generates a private/public key pair,
// connects to the server, performs the handshake
// and returns a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	// Generate a key pair.
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	// Connect to the server.
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	// Get the server's public key.
	var spk [32]byte
	_, err = io.ReadFull(conn, spk[:])
	if err != nil {
		conn.Close()
		return nil, err
	}

	// Send our public key to the server.
	_, err = conn.Write((*pub)[:])
	if err != nil {
		conn.Close()
		return nil, err
	}

	// Return a net.Conn wrapped in a SecureReadWriter for "secure" communication.
	return &SecureReadWriter{
		conn,
		NewSecureReader(conn, priv, &spk),
		NewSecureWriter(conn, priv, &spk),
	}, nil
}

// Serve starts a "secure" echo server on the given listener.
func Serve(l net.Listener) error {
	// Generate a key pair.
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}

	// Accept client connections.
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		// Handle a client connection.
		go func(conn net.Conn) {
			defer conn.Close()

			// Send our public key to the client.
			_, err := conn.Write((*pub)[:])
			if err != nil {
				return
			}

			// Get the client's public key.
			var cpk [32]byte
			_, err = io.ReadFull(conn, cpk[:])
			if err != nil {
				return
			}

			// Prepare reader and writer for "secure" communication.
			r := NewSecureReader(conn, priv, &cpk)
			w := NewSecureWriter(conn, priv, &cpk)
			buf := make([]byte, 32*1024)

			// Read messages and respond to them.
			for {
				// Read a message.
				n, err := r.Read(buf)
				if err != nil {
					return
				}

				// Echo the message back to the client.
				_, err = w.Write(buf[:n])
				if err != nil {
					return
				}
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
	if err != nil && err != io.EOF {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", buf[:n])
}
