package main

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

// Wrapper around an io.Reader that contains sharedKey for NaCl
type SecureReader struct {
	sharedKey [32]byte
	reader    io.Reader
}

// Implementation of io.Reader for SecureReader
// Messages are read and decrypted using the previous initialized keys
func (sr SecureReader) Read(p []byte) (n int, err error) {
	// Get the new nonce
	var nonce [24]byte
	n, err = sr.reader.Read(nonce[:])
	if err != nil {
		return 0, err
	}

	// Followed by length of enc message
	var msgLen uint64
	err = binary.Read(sr.reader, binary.BigEndian, &msgLen)
	if err != nil {
		return 0, err
	}

	// grab encrypted box
	encBox := make([]byte, msgLen)
	n, err = sr.reader.Read(encBox)
	if err != nil {
		return 0, err
	}

	// Read directly into the output slice p[:0]
	ret, valid := box.OpenAfterPrecomputation(p[:0], encBox, &nonce, &sr.sharedKey)
	if !valid {
		return 0, errors.New("Invalid encryption")
	}

	return len(ret), nil
}

// NewSecureReader instantiates a new SecureReader using given keys
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	sr := SecureReader{
		reader: r,
	}
	box.Precompute(&sr.sharedKey, pub, priv)
	return sr
}

// Wrapped around an io.Writer that implements NaCl encryption
type SecureWriter struct {
	sharedKey [32]byte
	writer    io.Writer
	nonce     [24]byte
	counter   uint64
}

func (sw SecureWriter) Write(p []byte) (n int, err error) {

	// output buffer will contain [nonce + msgLen + encMsg]
	outputBuffer := new(bytes.Buffer)

	// Update nonce filling the first 8 bytes with counter
	// Counter ensures nonce uniqueness and will take hundreds of years to overflow at uint64
	binary.BigEndian.PutUint64(sw.nonce[:], sw.counter)
	sw.counter++

	// Write nonce
	outputBuffer.Write(sw.nonce[:])

	ret := box.SealAfterPrecomputation(nil, p, &sw.nonce, &sw.sharedKey)

	// Preface encrypted message with length of that message
	binary.Write(outputBuffer, binary.BigEndian, uint64(len(ret)))

	// copy encrypted message to output buffer
	n, err = outputBuffer.Write(ret)
	if err != nil {
		return n, err
	}

	// flush output buffer to underlying writer
	outputBuffer.WriteTo(sw.writer)

	// caller only needs to know how many if it's bytes were written, not the encrypted byte count
	return len(p), nil
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	sw := SecureWriter{
		writer: w,
	}

	// Generate random starting nonce, we're going to write a sequence into it as well per message for uniqueness
	rand.Read(sw.nonce[:])

	box.Precompute(&sw.sharedKey, pub, priv)
	return sw
}

// Wrapper around two secure streams providing io.ReadWriteCloser
type SecureConnection struct {
	conn net.Conn
	r    io.Reader
	w    io.Writer
}

func (sc SecureConnection) Read(p []byte) (n int, err error) {
	return sc.r.Read(p)
}

func (sc SecureConnection) Write(p []byte) (n int, err error) {
	return sc.w.Write(p)
}

func (sc SecureConnection) Close() error {
	return sc.conn.Close()
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
		return nil, err
	}

	// Begin handshake
	// 1) Server sends public key
	// 2) Send local public key to server

	// Server sends public key, basic verification
	var serverPub [32]byte
	n, err := conn.Read(serverPub[:])
	if n != 32 || err != nil {
		return nil, err
	}

	// Send our public key to server, basic verification
	n, err = conn.Write(pub[:])
	if n != 32 || err != nil {
		return nil, err
	}

	sc := SecureConnection{
		conn: conn,
		w:    NewSecureWriter(conn, priv, &serverPub),
		r:    NewSecureReader(conn, priv, &serverPub),
	}

	return sc, nil
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

		go func(c net.Conn) {
			defer c.Close()

			// Begin handshake
			// 1) Send public key
			// 2) Get client public key

			// Immediately send server public key to client
			n, err := c.Write(pub[:])
			if n != 32 || err != nil {
				return
			}

			// client pub
			var clientPub [32]byte
			n, err = c.Read(clientPub[:])
			if n != 32 || err != nil {
				return
			}

			// Technically, the server doesn't even need to decrypt since the client is just echoing
			io.Copy(NewSecureWriter(conn, priv, &clientPub), NewSecureReader(conn, priv, &clientPub))
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
