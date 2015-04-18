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

const nonceLen = 24

var (
	// ErrKeyExchangeFailed is the error returned by Dial
	// and Serve if the key exchange fails.
	ErrKeyExchangeFailed = errors.New("key exchange failed")
	// ErrEncryptFailed is the error returned when a secure
	// writer is unable to encrypt a message.
	ErrEncryptFailed = errors.New("encrypt failed")
	// ErrDecryptFailed is the error returned when a secure
	// reader is unable to decrypt a message.
	ErrDecryptFailed = errors.New("decrypt failed")
)

type secureReader struct {
	r   io.Reader
	key [32]byte
	buf []byte
}

type secureWriter struct {
	w   io.Writer
	key [32]byte
	buf []byte
}

// Read reads from the underlying reader, decrpyting the
// data before storing it in p.
func (sr secureReader) Read(p []byte) (int, error) {
	// read the nonce and the length of the encrypted message first
	var nonce [24]byte
	_, err := io.ReadFull(sr.r, nonce[:])
	if err != nil {
		return 0, err
	}

	var length uint16
	if err := binary.Read(sr.r, binary.BigEndian, &length); err != nil {
		return 0, err
	}

	// make sure we have room to read in the encrypted data
	total := int(length + box.Overhead)
	if cap(sr.buf) < total {
		sr.buf = make([]byte, total, total*2)
	} else {
		sr.buf = sr.buf[:total]
	}

	// read in the data and decrypt it
	n, err := io.ReadFull(sr.r, sr.buf)
	if err != nil {
		return n, err
	}
	var ok bool
	msg, ok := box.OpenAfterPrecomputation(nil, sr.buf, &nonce, &sr.key)
	if !ok {
		return n, ErrDecryptFailed
	}
	if len(msg) != int(length) {
		return n, ErrDecryptFailed
	}
	return copy(p, msg), nil
}

// Write encrypts the data in p before writing it to the
// underlying writer.
func (sw secureWriter) Write(p []byte) (int, error) {
	// generate random nonce
	var nonce [nonceLen]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return 0, err
	}

	// write the plain-text nonce first, followed by len(p)
	if _, err := sw.w.Write(nonce[:]); err != nil {
		return 0, err
	}

	// we can assume a max message size of 32 kB, so a uint16 will
	// be sufficient to hold the length
	lenp := uint16(len(p))
	if err := binary.Write(sw.w, binary.BigEndian, lenp); err != nil {
		return 0, err
	}

	// "clear" the write buffer, encrypt, and send
	sw.buf = sw.buf[:0]
	sw.buf = box.SealAfterPrecomputation(sw.buf, p, &nonce, &sw.key)
	n, err := sw.w.Write(sw.buf)

	// seal adds box.Overhead additional bytes to p
	// but we don't want to report that we wrote more bytes than requested
	if n == len(p)+box.Overhead {
		n = len(p)
	}
	return n, err
}

// NewSecureReader instantiates a new secure Reader.
// The reader is capable of reading data written by a secure
// Writer with the specified public key.
func NewSecureReader(r io.Reader, priv, writerPub *[32]byte) io.Reader {
	sr := secureReader{r: r}
	box.Precompute(&sr.key, writerPub, priv)
	return sr
}

// NewSecureWriter instantiates a new secure Writer.
// The writer secures the data so that it can only be
// read by a secure Reader with the specified public key.
func NewSecureWriter(w io.Writer, priv, readerPub *[32]byte) io.Writer {
	sw := secureWriter{w: w}
	box.Precompute(&sw.key, readerPub, priv)
	return sw
}

type readWriteCloser struct {
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
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	// wait for the server's public key
	var serverKey [32]byte
	_, err = io.ReadFull(conn, serverKey[:])
	if err != nil {
		conn.Close()
		return nil, ErrKeyExchangeFailed
	}

	// then send our public key
	_, err = conn.Write(pub[:])
	if err != nil {
		conn.Close()
		return nil, ErrKeyExchangeFailed
	}

	return readWriteCloser{
		NewSecureReader(conn, priv, &serverKey),
		NewSecureWriter(conn, priv, &serverKey),
		conn,
	}, nil
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		// send our public key
		_, err = conn.Write(pub[:])
		if err != nil {
			conn.Close()
			return err
		}

		// get the client's public key
		var clientKey [32]byte
		_, err = io.ReadFull(conn, clientKey[:])
		if err != nil {
			conn.Close()
			return err
		}

		go func() {
			buf := make([]byte, 4096)
			reader := NewSecureReader(conn, priv, &clientKey)
			writer := NewSecureWriter(conn, priv, &clientKey)
			for {
				n, err := reader.Read(buf)
				if err != nil {
					conn.Close()
					return
				}
				// echo the data back
				writer.Write(buf[:n])
			}
		}()
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
