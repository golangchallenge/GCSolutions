package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"

	crand "crypto/rand"

	"golang.org/x/crypto/nacl/box"
)

// lengths of nonce and priv/pub key
const nonceLength = 24
const keyLength = 32

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, pub *[keyLength]byte) io.Reader {
	return &secureReader{r, priv, pub, bytes.Buffer{}}
}

type secureReader struct {
	r         io.Reader
	priv, pub *[keyLength]byte

	// Buffer with plain-text data
	buf bytes.Buffer
}

func (sr *secureReader) Read(p []byte) (n int, err error) {
	// If we have some plain text left, return it
	if sr.buf.Len() > 0 {
		return sr.buf.Read(p)
	}

	// read nonce for next message
	var nonce [nonceLength]byte
	_, err = io.ReadFull(sr.r, nonce[:])
	if err != nil {
		return 0, err
	}

	// read length of the encrypted message
	var l int16
	err = binary.Read(sr.r, binary.BigEndian, &l)
	if err != nil {
		return 0, err
	}

	cipherText := make([]byte, int(l))
	_, err = io.ReadFull(sr.r, cipherText)
	if err != nil {
		return 0, err
	}

	// let's decrypt cipherText
	plainText, ok := box.Open(nil, cipherText, &nonce, sr.pub, sr.priv)
	if !ok {
		return 0, fmt.Errorf("decrypt failed")
	}

	// Put plainText into buffer and let the read continue from there
	if len(plainText) == 0 {
		// this should not happen, since we never encrypt empty messages
		return 0, fmt.Errorf("empty decrypted message")
	}

	sr.buf.Write(plainText)
	return sr.buf.Read(p)
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[keyLength]byte) io.Writer {
	nonce := rand.Uint32()
	return &secureWriter{w, nonce, priv, pub}
}

type secureWriter struct {
	w         io.Writer
	lastNonce uint32
	priv, pub *[keyLength]byte
}

// Each Write() will create new Sealed message.
func (sw *secureWriter) Write(p []byte) (nn int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	sw.lastNonce = sw.lastNonce + 1
	var nonce [nonceLength]byte
	binary.BigEndian.PutUint32(nonce[:], sw.lastNonce)

	out := box.Seal(nil, p, &nonce, sw.pub, sw.priv)

	// write nonce
	err = binary.Write(sw.w, binary.BigEndian, nonce)
	if err != nil {
		return 0, err
	}

	// write length of the next message
	err = binary.Write(sw.w, binary.BigEndian, int16(len(out)))
	if err != nil {
		return 0, err
	}

	// write encrypted message itself
	_, err = sw.w.Write(out)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	pub, priv, err := box.GenerateKey(crand.Reader)
	if err != nil {
		return nil, err
	}

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	peerPub, err := exchangeKeys(conn, pub)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to exchange keys: %s", err)
	}

	r := NewSecureReader(conn, priv, peerPub)
	w := NewSecureWriter(conn, priv, peerPub)

	type rwc struct {
		io.Reader
		io.Writer
		io.Closer
	}

	return &rwc{r, w, conn}, nil
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	pub, priv, err := box.GenerateKey(crand.Reader)
	if err != nil {
		return err
	}

	for {
		c, err := l.Accept()
		if c != nil {
			go runEcho(c, pub, priv)
		}
		if err != nil {
			return err
		}
	}
}

// Does very simple public-key exchange, returns peer pubkey
func exchangeKeys(c net.Conn, pub *[keyLength]byte) (*[keyLength]byte, error) {
	// send our public key
	_, err := c.Write(pub[:])
	if err != nil {
		return nil, err
	}

	// read peer public key
	var peerPub [keyLength]byte
	_, err = io.ReadFull(c, peerPub[:])
	if err != nil {
		return nil, err
	}

	return &peerPub, nil
}

func runEcho(c net.Conn, pub, priv *[keyLength]byte) {
	defer c.Close()

	peerPub, err := exchangeKeys(c, pub)
	if err != nil {
		log.Println("error exchanging keys:", err)
		return
	}

	r := NewSecureReader(c, priv, peerPub)
	w := NewSecureWriter(c, priv, peerPub)

	// simply copy incoming data back to the connection
	_, err = io.Copy(w, r)
	if err != nil {
		log.Println("server error:", err)
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
