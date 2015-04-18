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

// maxMsgLen is the size of the per-connection message buffer in the echo server.
// because messages are limited to 32K:
// "We consider that our messages will always be smaller than 32KB"
const maxMsgLen = 32768

// secureReader implements a NACL Reader.
type secureReader struct {
	r      io.Reader
	priv   *[32]byte // receiver's secret key
	pub    *[32]byte // sender's public key
	shared [32]byte  // precomputed shared key
}

// NewSecureReader instantiates a new SecureReader.
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	sr := &secureReader{r: r, priv: priv, pub: pub}
	box.Precompute(&sr.shared, pub, priv)
	return sr
}

// Read verifies and decrypts a ciphertext using NACL.
// Wire format:
//  1. nonce (24 bytes)
//  2. size (2 bytes, little endian)
//  3. ciphertext (number of bytes given by previous value)
// Not sure if it would be worth beginning with a magic id.
// This simple implementation is equivalent to a ReadFull - the supplied buffer
// must be big enough to contain the entire plaintext (which is limited to 32K)
// as a successful read will return the entire message.
func (sr secureReader) Read(buf []byte) (int, error) {
	// read the 24 bytes nonce
	var nonce [24]byte
	if _, err := io.ReadFull(sr.r, nonce[:]); err != nil {
		return 0, errors.New("secureReader: cant read nonce: " + err.Error())
	}
	// read the ciphertext size
	var size uint16
	if err := binary.Read(sr.r, binary.LittleEndian, &size); err != nil {
		return 0, errors.New("secureReader: cant read ciphertext size: " + err.Error())
	}
	// check if the supplied buffer is big enough to contain the plain text
	if uint16(len(buf)) < size-box.Overhead {
		return 0, fmt.Errorf("secureReader: ciphertext too long (%d, max = %d)", size, len(buf)+box.Overhead)
	}
	// read the ciphertext
	ctext := make([]byte, size)
	if _, err := io.ReadFull(sr.r, ctext); err != nil {
		return 0, errors.New("secureReader: cant read ciphertext: " + err.Error())
	}
	// verify + decrypt ciphertext into buf and return size of plaintext
	if buf, ok := box.OpenAfterPrecomputation(buf[0:0], ctext, &nonce, &sr.shared); ok {
		return len(buf), nil
	}
	return 0, errors.New("secureReader: error opening box")
}

// secureWrite implemented a NACL Writer.
type secureWriter struct {
	w      io.Writer
	priv   *[32]byte // sender's secret key
	pub    *[32]byte // receiver's public key
	shared [32]byte  // precomputed shared key
}

// NewSecureWriter instantiates a new SecureWriter.
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	sw := &secureWriter{w: w, priv: priv, pub: pub}
	box.Precompute(&sw.shared, pub, priv)
	return sw
}

// Write encrypts and authenticates a message using NACL.
func (sw secureWriter) Write(msg []byte) (int, error) {
	// obtain a random, 24 bytes nonce
	var nonce [24]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return 0, errors.New("secureWriter: cant generate random nonce: " + err.Error())
	}
	// write nonce
	if _, err := sw.w.Write(nonce[:]); err != nil {
		return 0, errors.New("secureWriter: cant write nonce: " + err.Error())
	}
	// encrypt + authenticate plaintext to ciphertext
	ctext := box.SealAfterPrecomputation(nil, msg, &nonce, &sw.shared)
	// write ciphertext length
	if err := binary.Write(sw.w, binary.LittleEndian, uint16(len(ctext))); err != nil {
		return 0, errors.New("secureWriter: cant write ciphertext size: " + err.Error())
	}
	// write ciphertext and return plaintext length on success
	if _, err := sw.w.Write(ctext); err != nil {
		return 0, errors.New("secureWriter: cant write ciphertext: " + err.Error())
	}
	return len(msg), nil
}

// secureReadWriteCloser implements a secure ReadWriteCloser using NACL.
type secureReadWriteCloser struct {
	io.Reader
	io.Writer
	io.Closer
}

// NewSecureReadWriteCloser creates a secure connection using NACL.
// It needs an established connection and a key pair.
// A handshake is done first in order to exchange the public keys.
func NewSecureReadWriteCloser(rwc io.ReadWriteCloser, priv, pub *[32]byte) (io.ReadWriteCloser, error) {
	peerpub, err := handshake(rwc, pub)
	if err != nil {
		return nil, errors.New("handshake failed: " + err.Error())
	}
	return &secureReadWriteCloser{
		NewSecureReader(rwc, priv, peerpub),
		NewSecureWriter(rwc, priv, peerpub),
		rwc,
	}, nil
}

// handshake sends our public key and returns the peer public key.
// The key exchange is done in plain text as allowed by the exercise.
func handshake(rw io.ReadWriter, pub *[32]byte) (*[32]byte, error) {
	// send our public key
	if _, err := rw.Write(pub[:]); err != nil {
		return nil, err
	}
	// receive peer public key
	var peerpub [32]byte
	if _, err := io.ReadFull(rw, peerpub[:]); err != nil {
		return nil, err
	}
	return &peerpub, nil
}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	// generate key pair
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, errors.New("Dial: cant generate keys: " + err.Error())
	}
	// connect to the server
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, errors.New("Dial: cant connect to server: " + err.Error())
	}
	// setup secure connection
	rwc, err := NewSecureReadWriteCloser(conn, priv, pub)
	if err != nil {
		conn.Close()
		return nil, errors.New("Dial: cant create secure connection: " + err.Error())
	}
	return rwc, nil
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	// generate key pair
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return errors.New("Serve: cant generate keys: " + err.Error())
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			return errors.New("Serve: Accept() failed: " + err.Error())
		}
		go func(c net.Conn) {
			// setup secure connection
			rwc, err := NewSecureReadWriteCloser(c, priv, pub)
			if err != nil {
				log.Printf("Serve: cant create secure connection: " + err.Error())
				c.Close()
				return
			}
			defer rwc.Close()
			// read message
			var buf [maxMsgLen]byte
			n, err := rwc.Read(buf[:])
			if err != nil {
				log.Printf("Serve: cant read message: " + err.Error())
				return
			}
			// write back message
			if _, err := rwc.Write(buf[:n]); err != nil {
				log.Printf("Serve: cant write message: " + err.Error())
				return
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
