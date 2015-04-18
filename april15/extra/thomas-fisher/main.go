package main

import (
	"bytes"
	"crypto/rand"
	"encoding/gob"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"golang.org/x/crypto/nacl/box"
)

func exchangeKeys(conn io.ReadWriter, publicKey []byte) (*[32]byte, error) {
	var peerPublicKey [32]byte

	if _, err := conn.Write(publicKey); err != nil {
		return nil, err
	}

	if _, err := io.ReadFull(conn, peerPublicKey[:]); err != nil {
		return nil, err
	}

	return &peerPublicKey, nil
}

// ErrDecryptionFailed indicates problem decrypting an encrypted message - perhaps an invalid public/private key or nonce
var ErrDecryptionFailed = errors.New("decryption error")

type decrypter struct {
	reader     io.Reader
	privateKey *[32]byte
	publicKey  *[32]byte
}

type encrypter struct {
	writer     io.Writer
	privateKey *[32]byte
	publicKey  *[32]byte
}

type encryptedMessage struct {
	Data  []byte
	Nonce *[24]byte
}

// Write encrypts and writes data to the wrapped io.Writer
func (e encrypter) Write(p []byte) (int, error) {
	var nonce [24]byte
	var buffer bytes.Buffer

	if _, err := rand.Read(nonce[:]); err != nil {
		return 0, err
	}

	encrypted := box.Seal(nil, p, &nonce, e.publicKey, e.privateKey)
	msg := encryptedMessage{Data: encrypted, Nonce: &nonce}

	if err := gob.NewEncoder(&buffer).Encode(msg); err != nil {
		return 0, err
	}

	if _, err := buffer.WriteTo(e.writer); err != nil {
		return 0, nil
	}
	return len(p), nil
}

// Read decrypts an encrypted message from the underlying io.Reader
func (d decrypter) Read(p []byte) (int, error) {
	var msg encryptedMessage

	if len(p) == 0 {
		return 0, nil
	}

	if err := gob.NewDecoder(d.reader).Decode(&msg); err != nil {
		return 0, err
	}

	decrypted, ok := box.Open(nil, msg.Data, msg.Nonce, d.publicKey, d.privateKey)
	if !ok {
		return 0, ErrDecryptionFailed
	}

	n := copy(p, decrypted)
	return n, nil
}

// GenerateKeyPair returns a public and private key for use with NewSecureReader/NewSecureWriter
func GenerateKeyPair() (*[32]byte, *[32]byte, error) {
	return box.GenerateKey(rand.Reader)
}

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(reader io.Reader, privateKey, publicKey *[32]byte) io.Reader {
	return &decrypter{reader, privateKey, publicKey}
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(writer io.Writer, privateKey, publicKey *[32]byte) io.Writer {
	return &encrypter{writer, privateKey, publicKey}
}

// NewSecureReadWriteCloser wraps an existing readWriteCloser with encryption and decryption
func NewSecureReadWriteCloser(readWriteCloser io.ReadWriteCloser, privateKey, peerPublicKey *[32]byte) io.ReadWriteCloser {
	return struct {
		io.Reader
		io.Writer
		io.Closer
	}{
		NewSecureReader(readWriteCloser, privateKey, peerPublicKey),
		NewSecureWriter(readWriteCloser, privateKey, peerPublicKey),
		readWriteCloser,
	}
}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (conn io.ReadWriteCloser, err error) {
	var publicKey, privateKey, peerPublicKey *[32]byte

	publicKey, privateKey, err = GenerateKeyPair()
	if err != nil {
		return
	}

	conn, err = net.Dial("tcp", addr)
	if err != nil {
		return
	}

	peerPublicKey, err = exchangeKeys(conn, publicKey[:])
	if err != nil {
		conn.Close()
	} else {
		conn = NewSecureReadWriteCloser(conn, privateKey, peerPublicKey)
	}

	return
}

// Serve starts a secure echo server on the given listener.
// Upon new connections, it performs a key exchange
// and echoes any data it receives
func Serve(l net.Listener) error {
	var publicKey, privateKey, peerPublicKey *[32]byte

	publicKey, privateKey, err := GenerateKeyPair()
	if err != nil {
		return err
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			continue
		}

		go func(rawConnection net.Conn) {
			defer rawConnection.Close()
			peerPublicKey, err = exchangeKeys(conn, publicKey[:])
			if err != nil {
				return
			}

			encConn := NewSecureReadWriteCloser(conn, privateKey, peerPublicKey)

			// We consider that our messages will always be smaller than 32KB
			msg := make([]byte, 32000)
			n, err := encConn.Read(msg)
			if err != nil {
				return
			}
			encConn.Write(msg[:n])
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
