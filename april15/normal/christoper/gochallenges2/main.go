// Exercise solution code for gochallenge only

package main

import (
	"bytes"
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	"golang.org/x/crypto/nacl/box"
)

// go version go1.4.2 linux/amd64
// oddly, go does not complain these unused global variable
var randKey = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLKMNOPQRSTUVWXYZ")
var unused int

var nonceSize = 24
var errBufSize = errors.New("Buffer size not enough to encode/ decode")
var bufferSize = 1024

// NaclBuffer implement io Reader, writer and Conn
type NaclBuffer struct {
	w       io.Writer
	r       io.Reader
	keyout  []byte
	keyin   []byte
	pubkey  [32]byte
	privkey [32]byte
	//sharekey [32]byte
	conn net.Conn
}

// Write writes encrypted data to p
func (nacl *NaclBuffer) Write(p []byte) (n int, err error) {
	n = len(p)
	if n <= 0 {
		return 0, errBufSize
	}

	err = nacl.encrypt(p)
	if err != nil {
		return 0, nil
	}

	if _, ok := nacl.conn.(net.Conn); ok {
		n, err = nacl.conn.Write(nacl.keyout)
	} else {
		n, err = nacl.w.Write(nacl.keyout)
	}

	return n, err
}

// Read reads decrypted data to p
func (nacl *NaclBuffer) Read(p []byte) (n int, err error) {
	local := make([]byte, bufferSize)
	if _, ok := nacl.conn.(net.Conn); ok {
		n, err = nacl.conn.Read(local)
	} else {
		n, err = nacl.r.Read(local)
	}

	if n <= 0 {
		return 0, errBufSize
	}

	err = nacl.decrypt(local[0:n])
	if err != nil {
		return 0, err
	}

	p = p[:len(nacl.keyout)]
	n = copy(p, nacl.keyout)

	return n, nil
}

// Close derived from Conn interface
func (nacl *NaclBuffer) Close() error {
	return nacl.conn.Close()
}

// LocalAddr derived from Conn interface
func (nacl *NaclBuffer) LocalAddr() net.Addr {
	return nacl.conn.LocalAddr()
}

// RemoteAddr derived from Conn interface
func (nacl *NaclBuffer) RemoteAddr() net.Addr {
	return nacl.conn.RemoteAddr()
}

// SetDeadline derived from Conn interface
func (nacl *NaclBuffer) SetDeadline(t time.Time) error {
	return nacl.conn.SetDeadline(t)
}

// SetReadDeadline derived from Conn interface
func (nacl *NaclBuffer) SetReadDeadline(t time.Time) error {
	return nacl.conn.SetReadDeadline(t)
}

// SetWriteDeadline derived from Conn interface
func (nacl *NaclBuffer) SetWriteDeadline(t time.Time) error {
	return nacl.conn.SetWriteDeadline(t)
}

func (nacl *NaclBuffer) encrypt(m []byte) (err error) {
	mSize := len(m)
	if mSize <= 0 {
		return errBufSize
	}

	nacl.keyin = append(nacl.keyin, m[0:mSize]...)

	nonce := genNonce()

	//encrypted bytes are return, not stored in out
	data := box.Seal(nil, nacl.keyin, &nonce, &nacl.pubkey, &nacl.privkey)
	//append nonce to msg data
	nacl.keyout = append(data, nonce[:]...)

	nacl.keyin = nil
	return nil
}

func (nacl *NaclBuffer) decrypt(m []byte) (err error) {
	mSize := len(m)
	if mSize <= nonceSize {
		return errBufSize
	}
	var nonce [24]byte
	nacl.keyin = append(nacl.keyin, m[0:mSize-nonceSize]...)
	copy(nonce[:], m[mSize-nonceSize:mSize])

	var out []byte
	data, _ := box.Open(out, nacl.keyin, &nonce, &nacl.pubkey, &nacl.privkey)
	nacl.keyout = data

	nacl.keyin = nil
	return nil
}

func genNonce() [24]byte {
	var b [24]byte
	_, err := rand.Read(b[:])
	if err != nil {
		log.Fatal("Error Generating Nonce : ", err)
	}
	return b
}

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	var nacl = new(NaclBuffer)
	nacl.privkey = *priv
	nacl.pubkey = *pub
	nacl.r = r
	return nacl
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	var nacl = new(NaclBuffer)
	nacl.privkey = *priv
	nacl.pubkey = *pub
	nacl.w = w
	return nacl
}

// NewSecureRW : wrapping conn
func NewSecureRW(rw net.Conn, priv, pub *[32]byte) net.Conn {
	var nacl = new(NaclBuffer)
	nacl.privkey = *priv
	nacl.pubkey = *pub
	nacl.conn = rw
	return nacl
}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		log.Fatal("Address Resolve Failed ", err.Error())
		return nil, err
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return nil, err
	}

	//Generating keys
	cpubKey, cprivKey, err := box.GenerateKey(genRandomBytes())

	if err != nil {
		log.Fatal("Client : key generation failed ", err.Error())
		return nil, err
	}

	if _, err := conn.Write(cpubKey[:]); err != nil {
		return nil, err
	}

	var cspubKey [32]byte

	key := make([]byte, 32)
	i, err := conn.Read(key)
	if err != nil {
		return nil, err
	}
	copy(cspubKey[:], key[0:i])

	io := NewSecureRW(conn, cprivKey, &cspubKey)
	return io, nil
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		defer conn.Close()
		//exchange key
		var scpubKey [32]byte
		cKey := make([]byte, 32)
		i, err := conn.Read(cKey)
		if err != nil {
			return err
		}
		copy(scpubKey[:], cKey[0:i])

		spubKey, sprivKey, err := box.GenerateKey(genRandomBytes())
		if err != nil {
			log.Fatal("Server: Key generation failed ", err.Error())
			return err
		}

		if _, err := conn.Write(spubKey[:]); err != nil {
			return err
		}

		conn = NewSecureRW(conn, sprivKey, &scpubKey)
		//go handleConn(conn)
		handleConn(conn)
	}
}

func handleConn(conn net.Conn) error {
	buf := make([]byte, bufferSize)
	n, err := conn.Read(buf)
	_, err = conn.Write(buf[0:n])
	return err
}

func genRandomBytes() io.Reader {
	var buf [256]byte
	rand.Read(buf[:])
	return bytes.NewReader(buf[:])
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
