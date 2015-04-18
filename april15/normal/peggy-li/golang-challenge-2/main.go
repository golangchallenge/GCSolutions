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
	"golang.org/x/crypto/nacl/secretbox"
)

// the number of bytes in a nonce
const nonceSize = 24

// the number of bytes in a public/private key
const keySize = 32

// SecureReader is a secure reader.
type SecureReader struct {
	reader     io.Reader
	privateKey *[32]byte
	publicKey  *[32]byte
}

// Read a decrypted message into p.
//
// The SecureReader reads the contents written by a SecureWriter.
// It will attempt to read/parse according to the following format:
//
// - Nonce: nonceSize bytes (in this case, 24)
// - Message Length (uint32): length of the encrypted message, in bytes.
// - Encrypted Message: the encrypted message itself.
//
// It decrypts the message using the nonce and the values of the
// SecureReader's public and private keys and copies the decrypted
// message into the first n bytes of p.
//
// Returns the length (in bytes) of the decrypted message.
func (r SecureReader) Read(p []byte) (n int, err error) {
	// read the nonce
	var nonce [nonceSize]byte
	err = binary.Read(r.reader, binary.LittleEndian, &nonce)
	if err != nil && err != io.EOF {
		return 0, errors.New("could not read nonce: " + err.Error())
	}

	// read the length of the encrypted message
	var msgLen uint32
	err = binary.Read(r.reader, binary.LittleEndian, &msgLen)
	if err != nil && err != io.EOF {
		return 0, errors.New("could not read message length: " + err.Error())
	}

	// read the encrypted message itself
	encrypted := make([]byte, msgLen)
	err = binary.Read(r.reader, binary.LittleEndian, &encrypted)
	if err != nil && err != io.EOF {
		return 0, errors.New("could not read encrypted message: " + err.Error())
	}

	// decrypt the message
	decrypted, res := box.Open(nil, encrypted, &nonce, r.publicKey, r.privateKey)
	if res == false {
		return 0, errors.New("error decrypting mesage with box.Open")
	}

	copy(p, decrypted)

	return int(msgLen) - secretbox.Overhead, nil
}

// SecureWriter is a secure writer.
type SecureWriter struct {
	writer     io.Writer
	privateKey *[32]byte
	publicKey  *[32]byte
}

// Write the message (contents of p) to w's underlying data stream.
//
// The message is first encrypted using a randomly generated nonce
// and the SecureWriter's public and private keys.
//
// The total fields written to the SecureWriter's data stream are:
//
// - Nonce: randomly generated nonce, nonceSize bytes (in this case, 24)
// - Message Length (uint32): length of the encrypted message, in bytes.
// - Encrypted Message: the encrypted message itself.
//
// All values use LittleEndian encoding by convention.
//
// Returns the length (in bytes) of the box (encrypted message).
// This does not include the sizes of the nonce or message length.
func (w SecureWriter) Write(p []byte) (n int, err error) {
	nonce, err := generateNonce()
	if err != nil {
		return 0, errors.New("could not generate nonce: " + err.Error())
	}

	// encrypt the message
	box := box.Seal(nil, p, nonce, w.publicKey, w.privateKey)

	// write the nonce
	if err = binary.Write(w.writer, binary.LittleEndian, nonce); err != nil {
		return 0, errors.New("could not write nonce: " + err.Error())
	}

	// write the encrypted message length
	boxLen := uint32(len(box))
	if err = binary.Write(w.writer, binary.LittleEndian, boxLen); err != nil {
		return 0, errors.New("could not write box length: " + err.Error())
	}

	// write the message itself
	if err = binary.Write(w.writer, binary.LittleEndian, box); err != nil {
		return 0, errors.New("could not write encrypted message: " + err.Error())
	}

	return int(boxLen), nil
}

// SecureReadWriteCloser is a wrapper around a connection.
// It allows Dial to return an io.ReadWriteCloser that uses
// SecureReader's Read and SecureWriter's Write implementations.
type SecureReadWriteCloser struct {
	reader SecureReader
	writer SecureWriter
	conn   net.Conn
}

// Read from p using the SecureReadWriteCloser's SecureReader.
func (s SecureReadWriteCloser) Read(p []byte) (n int, err error) {
	return s.reader.Read(p)
}

// Write to p using the SecureReadWriteCloser's SecureWriter.
func (s SecureReadWriteCloser) Write(p []byte) (n int, err error) {
	return s.writer.Write(p)
}

// Close the SecureReadWriteCloser's underlying connection.
func (s SecureReadWriteCloser) Close() error {
	return s.conn.Close()
}

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	return SecureReader{r, priv, pub}
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	return SecureWriter{w, priv, pub}
}

// NewSecureReadWriteCloser instantiates a new SecureReadWriteCloser
func NewSecureReadWriteCloser(c net.Conn, priv, pub *[32]byte) io.ReadWriteCloser {
	return SecureReadWriteCloser{
		SecureReader{reader: c, privateKey: priv, publicKey: pub},
		SecureWriter{writer: c, privateKey: priv, publicKey: pub},
		c,
	}
}

// generateNonce returns a randomly generated nonce of nonceSize bytes.
func generateNonce() (*[nonceSize]byte, error) {
	nonce := [nonceSize]byte{}
	if _, err := rand.Read(nonce[:]); err != nil {
		return nil, err
	}
	return &nonce, nil
}

// Dial generates a private/public key pair,
// connects to the server, performs the handshake,
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	// generate public/private key pair
	publicKey, privateKey, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	// send own public key to the server
	// NOTE|TODO: The keys are currently sent in plain text,
	// which is okay for the challenge, but ideally this should use
	// a more secure algorithm like Diffie-Hellman to get a shared key.
	if _, err := conn.Write(publicKey[:]); err != nil {
		log.Fatal("error sending public key: " + err.Error())
	}

	// get the server's public key
	peerPublicKey := [keySize]byte{}
	if _, err = conn.Read(peerPublicKey[:]); err != nil {
		return nil, err
	}

	return NewSecureReadWriteCloser(conn, privateKey, &peerPublicKey), nil
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	defer l.Close()

	// generate a public/private key pair
	publicKey, privateKey, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return errors.New("could not generate public/private keys: " + err.Error())
	}

	for {
		// accept incoming connection
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		go func(c net.Conn) {
			defer c.Close()

			// get the client's public key
			peerPublicKey := [keySize]byte{}
			if _, err := c.Read(peerPublicKey[:]); err != nil {
				log.Fatal("error in key exchange: " + err.Error())
			}

			// send own public key to the client
			if _, err := c.Write(publicKey[:]); err != nil {
				log.Fatal("error sending public key: " + err.Error())
			}

			// create a new SecureReadWriteCloser for the connection
			NewSecureReadWriteCloser(c, privateKey, &peerPublicKey)

			io.Copy(c, c)
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
