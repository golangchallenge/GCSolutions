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

func init() {
	sessions = make(map[[32]byte]*[24]byte)
}

const (
	publicKeySize  = 32
	privateKeySize = 32
	nonceSize      = 24
	payloadItem    = 'P'
	keyItem        = 'K'
)

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

// Map that holds nonces with clients identified by public key
var sessions map[[32]byte]*[24]byte

// Sets the nonce shared between dialer and server; either increments the nonce if the nonce exists or sets a random nonce if one does not exists
func setSession(pub *[32]byte) {
	if _, exists := sessions[*pub]; !exists {
		var n [24]byte
		io.ReadFull(rand.Reader, n[:])
		sessions[*pub] = &n
	} else {
		incrementNonce(sessions[*pub])
	}
}

// Increment nonce by 1; once max has been reached, the 24 byte nonce will be set to 0
func incrementNonce(nonce *[24]byte) {
	var maxUint64 uint64 = 1<<64 - 1
	left, _ := binary.Uvarint(nonce[:8])
	middle, _ := binary.Uvarint(nonce[8:16])
	right, _ := binary.Uvarint(nonce[16:])
	if right < maxUint64 {
		right++
	} else {
		right = 0
		if middle < maxUint64 {
			middle++
		} else {
			middle = 0
			if left < maxUint64 {
				left++
			} else {
				left = 0
			}
		}
	}
	binary.PutUvarint(nonce[16:], right)
	binary.PutUvarint(nonce[8:16], middle)
	binary.PutUvarint(nonce[:8], left)
}

// Common insecure reader; encapsulates typical insecure read
func readRaw(conn io.Reader, n int) ([]byte, error) {
	buf := make([]byte, n)
	n, err := (conn).Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

// Used in key exchange; returns either nonce, or public key depending on itemName (keyItem or payloadItem constants)
func extractKeyOrNonce(dst, src []byte, itemName rune) (n int) {
	switch itemName {
	case keyItem:
		n = copy(dst, src[:publicKeySize])
	case payloadItem:
		n = copy(dst, src[publicKeySize:])
	}
	return
}

// Appends nonce to public key
func constructKeyExchange(pub *[32]byte, nonce []byte) []byte {
	var b []byte
	b = append(b, pub[:]...)
	b = append(b, nonce...)
	return b
}

// Handles key exchange for Dial; sends public key and nonce to Server; receives server's public key, establishes secure connection
func dialKeyExchange(conn *net.Conn, prot, addr string) (rwc io.ReadWriteCloser, err error) {
	//generate key pair for client
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return
	}

	var prenonce [24]byte
	io.ReadFull(rand.Reader, prenonce[:])
	b := constructKeyExchange(pub, prenonce[:])
	_, err = (*conn).Write(b)
	if err != nil {
		return
	}
	//read s_init from server
	b, err = readRaw(*conn, 512)
	if err != nil {
		return
	}
	var peerPub [32]byte
	eNonce := make([]byte, 512)
	extractKeyOrNonce(peerPub[:], b, keyItem)
	n := extractKeyOrNonce(eNonce, b, payloadItem)
	eNonce = eNonce[:n]
	var nonce [24]byte
	copy(nonce[:], prenonce[:])
	sessions[peerPub] = &nonce

	//create secure connection using public key
	rwc = NewSecureReadWriteCloser(conn, &peerPub, priv)

	return
}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	secureConn, err := dialKeyExchange(&conn, "tcp", addr)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return secureConn, err
}

// Reads public key sent from Dial; extracts dialer public key
// and nonce and stores both in the sessions map;
// sends public key back to dialer
func serveKeyExchange(conn *net.Conn, priv, pub *[32]byte, clientKeyExchange []byte) (srwc io.ReadWriteCloser, err error) {
	var (
		peerPub [32]byte
		nonce   [24]byte
	)
	extractKeyOrNonce(peerPub[:], clientKeyExchange, keyItem)
	extractKeyOrNonce(nonce[:], clientKeyExchange, payloadItem)
	sessions[peerPub] = &nonce
	_, err = (*conn).Write(pub[:])
	if err != nil {
		return
	}
	srwc = NewSecureReadWriteCloser(conn, &peerPub, priv)
	return
}

// Handles connection from Serve function accept loop;
// establishes secure connection and then receives and
// echos message from and to client on secure connection
func handleConnection(conn net.Conn, pub, priv *[32]byte) {
	defer conn.Close()
	b, err := readRaw(conn, 512)
	if err != nil {
		log.Println(err)
		return
	}
	if len(b) != (publicKeySize + nonceSize) {
		conn.Write([]byte("Invalid Key Format"))
		return
	}

	srwc, err := serveKeyExchange(&conn, priv, pub, b)
	if err != nil {
		log.Println(err)
		return
	}

	buf := make([]byte, 1024)
	n, err := srwc.Read(buf)
	buf = buf[:n]
	_, err = srwc.Write(buf)
	if err != nil {
		log.Println(err)
	}
}

//Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println(err)
			return err
		}
		handleConnection(conn, pub, priv)
	}
	return nil
}

// Secure Reader type, holds private key, server public key, nonce and the insecure reader
type SecureReader struct {
	PrivateKey, PeerPublicKey *[32]byte
	Nonce                     *[24]byte
	Reader                    *io.Reader
}

// Receives ciphertext and decrypts to plaintext into byte array b; returns bytes read (length of plaintext) and error
func (sr SecureReader) Read(b []byte) (n int, err error) {
	c, err := readRaw(*(sr.Reader), 512)
	if err != nil {
		return 0, err
	}
	p, ok := box.Open(nil, c, sr.Nonce, sr.PeerPublicKey, sr.PrivateKey)
	if !ok {
		err = errors.New("decrypt fail")
		return
	}
	n = len(p)
	p = p[:n]
	copy(b, p)
	return
}

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	setSession(pub)
	return SecureReader{PrivateKey: priv, PeerPublicKey: pub, Reader: &r, Nonce: sessions[*pub]}
}

// Secure Writer type, holds private key, dialer public key, nonce and insecure writer
type SecureWriter struct {
	PrivateKey, PeerPublicKey *[32]byte
	Nonce                     *[24]byte
	Writer                    *io.Writer
}

// Encrypts plaintext and writes cipher text to stream
func (sw SecureWriter) Write(b []byte) (n int, err error) {
	c := box.Seal(nil, b, sw.Nonce, sw.PeerPublicKey, sw.PrivateKey)
	n, err = (*sw.Writer).Write(c)
	if err != nil {
		log.Println(err)
	}
	return
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	setSession(pub)
	return SecureWriter{PrivateKey: priv, PeerPublicKey: pub, Writer: &w, Nonce: sessions[*pub]}
}

// ReadWriteCloser Interface encapsulates private and public keys with insecure connection
type SecureReadWriteCloser struct {
	privateKey, peerPublicKey *[32]byte
	conn                      *net.Conn
}

// Creates Secure Reader and uses its Read method
func (srwc SecureReadWriteCloser) Read(b []byte) (n int, err error) {
	r := NewSecureReader(*srwc.conn, srwc.privateKey, srwc.peerPublicKey)
	return r.Read(b)
}

// Creates Secure Writer and uses its Write method
func (srwc SecureReadWriteCloser) Write(b []byte) (n int, err error) {
	w := NewSecureWriter(*srwc.conn, srwc.privateKey, srwc.peerPublicKey)
	return w.Write(b)
}

// Closes connection
func (srwc SecureReadWriteCloser) Close() (err error) {
	return (*srwc.conn).Close()
}

// Creates SecureReadWriteCloser
func NewSecureReadWriteCloser(rwc *net.Conn, pub, priv *[32]byte) io.ReadWriteCloser {
	return SecureReadWriteCloser{
		privateKey:    priv,
		peerPublicKey: pub,
		conn:          rwc}
}
