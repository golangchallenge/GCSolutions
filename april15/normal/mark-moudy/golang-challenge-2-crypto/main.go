package main

// Golang Challenge Submission by Mark Moudy - April 06, 2015

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"golang.org/x/crypto/nacl/box"
	"io"
	"log"
	"net"
	"os"
)

const (
	keySize   = 32
	nonceSize = 24
)

// Metadata contains the details needed to open the message
type Metadata struct {
	// Length specifies the length of the sealed message
	Length int16
	// Nonce is the unique key for each message pair
	Nonce [nonceSize]byte
}

// A SecureReader implements io.Reader to decrypt messages.
type SecureReader struct {
	priv, pub *[keySize]byte
	reader    io.Reader
}

// NewSecureReader instantiates a new SecureReader to decrypt messages
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	return &SecureReader{
		priv:   priv,
		pub:    pub,
		reader: r,
	}
}

//Read decrypts a message using the packed metadata object and stored keys.
func (sr *SecureReader) Read(p []byte) (int, error) {
	md := &Metadata{}
	err := binary.Read(sr.reader, binary.LittleEndian, md)
	if err != nil {
		println(err)
		return 0, err
	}
	rec := make([]byte, md.Length)
	io.ReadFull(sr.reader, rec)

	msg, ok := box.Open(nil, rec, &md.Nonce, sr.pub, sr.priv)
	if !ok {
		return 0, errors.New("Unable to open message")
	}
	copy(p[:], msg)
	return len(msg), nil
}

// A SecureWriter implements io.Writer to encrypt messages.
type SecureWriter struct {
	priv, pub *[keySize]byte
	writer    io.Writer
}

// NewSecureWriter instantiates a new SecureWriter to encrypt messages
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	return &SecureWriter{
		priv:   priv,
		pub:    pub,
		writer: w,
	}
}

//Write prepends the metadata object to out which adds it to the front of the sealed message.
func (sw *SecureWriter) Write(p []byte) (int, error) {
	nonce, err := genNonce()
	if err != nil {
		println(err)
		return 0, err
	}
	md := &Metadata{
		Length: int16(len(p) + box.Overhead),
		Nonce:  *nonce,
	}
	msg := box.Seal(nil, p, &md.Nonce, sw.pub, sw.priv)
	buf := bytes.Buffer{}

	binary.Write(&buf, binary.LittleEndian, md)
	binary.Write(&buf, binary.LittleEndian, msg)
	sw.writer.Write(buf.Bytes())

	return binary.Size(msg), nil
}

func genNonce() (*[nonceSize]byte, error) {
	var nonce [nonceSize]byte
	n, err := rand.Read(nonce[:])
	if err != nil {
		return nil, err
	}

	if n != nonceSize {
		return nil, errors.New("Nonce doesn't contain 24 random bytes")
	}
	return &nonce, nil
}

// A SecureReadWriteCloser represents a pair of io.Reader and io.Writers to
// encrypt and decrypt messages.
type SecureReadWriteCloser struct {
	reader io.Reader
	writer io.Writer
}

// NewSecureReadWriteCloser returns a new SecureReadWriteCloser.
func NewSecureReadWriteCloser(conn net.Conn, priv, pub *[32]byte) io.ReadWriteCloser {
	return &SecureReadWriteCloser{
		NewSecureReader(conn, priv, pub),
		NewSecureWriter(conn, priv, pub),
	}
}

func (swc *SecureReadWriteCloser) Read(p []byte) (n int, err error) {
	return swc.reader.Read(p)
}

func (swc *SecureReadWriteCloser) Write(p []byte) (n int, err error) {
	return swc.writer.Write(p)
}

// Close exists to implement the io.ReadWriteCloser interface and only
// returns nil because data is in memory and we don't care about it.
func (swc *SecureReadWriteCloser) Close() error {
	return nil
}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatalf("Dial Error: %v\n", err)
		return nil, err
	}
	return NewSecureReadWriteCloser(conn, priv, pub), nil
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	for {
		//wait for connection
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		go func(c net.Conn) {
			defer c.Close()
			buf := make([]byte, 32000) //max expected message size of 32KB
			_, err := c.Read(buf)
			if err != nil {
				log.Fatalf("Serve Read Error: %v\n", err)
			}

			rec := make([]byte, binary.Size(buf))
			err = binary.Read(bytes.NewReader(buf), binary.LittleEndian, rec)
			if err != nil {
				println(err)
			}

			c.Write(rec)
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
