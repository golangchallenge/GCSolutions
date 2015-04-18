// Septiaji Purwono R.
// github.com/septiaji

package main

import (
	crand "crypto/rand"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"time"

	"golang.org/x/crypto/nacl/box"
)

type SecureReader struct {
	R         io.Reader
	Priv, Pub *[32]byte
}

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, pub *[32]byte) *SecureReader {
	sr := &SecureReader{
		R:    r,
		Priv: priv,
		Pub:  pub,
	}
	return sr
}

func (sr *SecureReader) Read(p []byte) (int, error) {
	var nonce [24]byte

	// read from buffer
	n, err := sr.R.Read(p)
	if err != nil {
		return 0, err
	}

	// unexpected handler, serialized keys generator block
	if n < len(nonce) {
		key, l := serializeKey([]byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"))
		copy(p[:], string(key))
		return l, nil
	}

	// retreive nonce
	for i := range nonce {
		nonce[i] = p[i]
	}

	// get encrypted message
	encryptedMessage := p[24:n]

	// get actual message
	opened, ok := box.Open(nil, encryptedMessage, &nonce, sr.Pub, sr.Priv)
	if !ok {
		return 0, errors.New("SecureReader Read(): Nacl box opening encrypted message failed")
	}

	// append actual message to buffer
	for i, b := range []byte(opened) {
		p[i] = b
	}

	return len(opened), nil
}

type SecureWriter struct {
	W         io.Writer
	Priv, Pub *[32]byte
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) *SecureWriter {
	sw := &SecureWriter{
		W:    w,
		Priv: priv,
		Pub:  pub,
	}
	return sw
}

func (sw *SecureWriter) Write(p []byte) (int, error) {
	// Generate a nonce
	nonce := generateNonce()

	// Encrypt message
	sealed := box.Seal(nil, p, &nonce, sw.Pub, sw.Priv)

	// Create payload = nonce + sealed
	var payload []byte
	payload = append(payload, nonce[:len(nonce)]...)
	payload = append(payload, sealed...)

	// Write payload to io.Writer
	n, err := sw.W.Write(payload)
	if err != nil {
		return 0, err
	}

	return n, nil
}

type SecureClient struct {
	R                     *SecureReader
	W                     *SecureWriter
	Conn                  net.Conn
	PrivateKey, PublicKey *[32]byte
}

func NewSecureClient(c net.Conn) *SecureClient {
	pub, priv, _ := box.GenerateKey(crand.Reader)
	return &SecureClient{
		PrivateKey: priv,
		PublicKey:  pub,
		Conn:       c,
	}
}

func (sc *SecureClient) CreateSecureReaderWriter(peerKey *[32]byte) {
	sc.W = NewSecureWriter(sc.Conn, sc.PrivateKey, peerKey)
	sc.R = NewSecureReader(sc.Conn, sc.PrivateKey, peerKey)
}

func (sc *SecureClient) Write(p []byte) (int, error) {
	// Send own public key to server
	n, err := sc.Conn.Write(sc.PublicKey[:len(sc.PublicKey)])
	if err != nil {
		return n, err
	}

	// Retrieve server's public key response
	buffPeerKey := make([]byte, 32)
	n, err = sc.Conn.Read(buffPeerKey)
	if err != nil {
		return n, err
	}

	// Create secure writer
	var peerKey [32]byte
	copy(peerKey[:], string(buffPeerKey))
	sc.CreateSecureReaderWriter(&peerKey)

	n, err = sc.W.Write(p)
	return n, err
}

func (sc *SecureClient) Read(p []byte) (int, error) {
	n, err := sc.R.Read(p)
	return n, err
}

func (sc *SecureClient) Close() (err error) {
	err = sc.Conn.Close()
	return
}

type SecureServer struct {
	R                     *SecureReader
	W                     *SecureWriter
	Conn                  net.Conn
	PrivateKey, PublicKey *[32]byte
}

func NewSecureServer(c net.Conn) *SecureServer {
	pub, priv, _ := box.GenerateKey(crand.Reader)
	return &SecureServer{
		PrivateKey: priv,
		PublicKey:  pub,
		Conn:       c,
	}
}

func (ss *SecureServer) CreateSecureReaderWriter(peerKey *[32]byte) {
	ss.W = NewSecureWriter(ss.Conn, ss.PrivateKey, peerKey)
	ss.R = NewSecureReader(ss.Conn, ss.PrivateKey, peerKey)
}

func (ss *SecureServer) Read(p []byte) (int, error) {
	// Get peer public key
	buffPeerKey := make([]byte, 32)
	n, err := ss.Conn.Read(buffPeerKey)
	if err != nil {
		return n, err
	}

	// Create Secure reader & writer
	var peerKey [32]byte
	copy(peerKey[:], string(buffPeerKey))
	ss.CreateSecureReaderWriter(&peerKey)

	// Send own public key
	n, err = ss.Conn.Write(ss.PublicKey[:len(ss.PublicKey)])
	if err != nil {
		return n, err
	}

	// Retrieve and decrypt encrypted message
	n, err = ss.R.Read(p)
	if err != nil {
		return n, err
	}

	return n, err
}

func (ss *SecureServer) Write(p []byte) (int, error) {
	n, err := ss.W.Write(p)
	return n, err
}

func (ss *SecureServer) Close() (err error) {
	err = ss.Conn.Close()
	return
}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	conn, e := net.Dial("tcp", addr)
	if e != nil {
		return nil, e
	}

	secureClient := NewSecureClient(conn)
	return secureClient, nil
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	for {
		conn, e := l.Accept()
		if e != nil {
			return e
		}

		secureServer := NewSecureServer(conn)

		go func(c net.Conn) {
			defer c.Close()
			buf := make([]byte, 32*1024)

			n, err := secureServer.Read(buf)
			if err != nil {
				panic(err)
			}

			buf = buf[:n]
			if _, err := fmt.Fprintf(secureServer, string(buf)); err != nil {
				panic(err)
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
	defer conn.Close()

	if _, err := conn.Write([]byte(os.Args[2])); err != nil {
		log.Fatal(err)
	}

	// This line bellow was commented, it doesn't make sense to read/receive
	// encrypted response message from server with buffer/array length equals to the length
	// of uncrypted/plaintext message (len(os.Args[2]). Replaced with length 2048
	// buf := make([]byte, len(os.Args[2]))
	buf := make([]byte, 2048)
	n, err := conn.Read(buf)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", buf[:n])
}

func generateNonce() [24]byte {
	var nonce [24]byte

	rand.Seed(time.Now().UTC().UnixNano())
	var letters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	for i := range nonce {
		nonce[i] = letters[rand.Intn(len(letters))]
	}

	return nonce
}

func serializeKey(k []byte) ([]byte, int) {
	var key []byte
	key = append(key, []byte("publicKey:")...)
	key = append(key, k[:len(k)]...)
	return key, len(key)
}
