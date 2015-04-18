package main

import (
	"crypto/rand"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	a "golang.org/x/crypto/nacl/box"
)

var (
	version       = "0.1"
	hdrVersion    = "Version"
	hdrPublicKey  = "Public-Key"
	hdrPrivateKey = "Private-Key"
	hdrNonce      = "Nonce"
)

type securedReader struct {
	R         io.Reader
	priv, pub *[32]byte
}

func (l *securedReader) Read(p []byte) (n int, err error) {
	data := make([]byte, 1024)
	//blocking reader
	for n == 0 {
		n, _ = l.R.Read(data)
	}
	b, _ := pem.Decode(data)
	if b == nil {
		return 0, errors.New("pem decoding failed")
	}
	bnonce, _ := b.Headers[hdrNonce]
	if len(bnonce) != 24 {
		return 0, errors.New("invalid nonce")
	}
	var nonce [24]byte
	copy(nonce[:], bnonce)
	dec, ok := a.Open(nil, b.Bytes, &nonce, l.pub, l.priv)
	if !ok {
		return 0, errors.New("Error in decryption")
	}
	copy(p, dec)
	n = len(dec)
	return
}

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	return &securedReader{r, priv, pub}
}

type securedWriter struct {
	R         io.Writer
	priv, pub *[32]byte
}

func (l *securedWriter) Write(p []byte) (n int, err error) {
	nonce := loadNonce()
	encr := a.Seal(nil, p, nonce, l.pub, l.priv)
	b := &pem.Block{
		Headers: map[string]string{
			hdrVersion:   version,
			hdrPublicKey: string(l.pub[:]),
			hdrNonce:     string(nonce[:]),
		},
		Bytes: encr,
	}
	n, err = l.R.Write(pem.EncodeToMemory(b))
	if err != nil {
		return 0, errors.New("Error while writing")
	}
	return
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	return &securedWriter{w, priv, pub}
}

type transport struct {
	w io.Writer
	r io.Reader
}

func (tr *transport) Write(p []byte) (n int, err error) {
	return tr.w.Write(p)
}
func (tr *transport) Read(p []byte) (n int, err error) {
	return tr.r.Read(p)
}
func (tr *transport) Close() error {
	//TODO might missing something here
	return nil
}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	priv, pub, err := a.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	handshakeC(conn, priv, pub)
	secureR := NewSecureReader(conn, priv, pub)
	secureW := NewSecureWriter(conn, priv, pub)
	return &transport{secureW, secureR}, nil
}

func handshakeC(conn net.Conn, priv, pub *[32]byte) {
	b := &pem.Block{
		Type: "INIT",
		Headers: map[string]string{
			hdrVersion:    version,
			hdrPublicKey:  string(pub[:]),
			hdrPrivateKey: string(priv[:]),
		},
	}
	conn.Write(pem.EncodeToMemory(b))
	buf := make([]byte, 1024)
	reqLen, err := conn.Read(buf)
	if err != nil {
		log.Fatal(err)
		return
	}
	bb, _ := pem.Decode(buf[:reqLen])
	if bb != nil && bb.Type == "INIT-ACK" {
	}
	return

}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleRequest(conn)
		break
	}
	return nil
}

func handshakeS(conn net.Conn) (priv *[32]byte, pub *[32]byte) {
	buf := make([]byte, 1024)
	reqLen, err := conn.Read(buf)
	if err != nil {
		log.Print(err)
		return
	}
	//ACK
	bb := &pem.Block{
		Type: "INIT-ACK",
		Headers: map[string]string{
			hdrVersion: version,
		},
	}
	conn.Write(pem.EncodeToMemory(bb))

	b, _ := pem.Decode(buf[:reqLen])
	if b != nil && b.Type == "INIT" {
		var priv [32]byte
		copy(priv[:], b.Headers[hdrPrivateKey])
		var pub [32]byte
		copy(pub[:], b.Headers[hdrPublicKey])
		return &priv, &pub
	}
	return
}

func handleRequest(conn net.Conn) {
	pub, priv := handshakeS(conn)
	tr := &transport{NewSecureWriter(conn, pub, priv),
		NewSecureReader(conn, pub, priv)}
	for {
		buf := make([]byte, 1024)
		reqLen, err := tr.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
			return
		}
		tr.Write(buf[:reqLen])
	}

	err := conn.Close()
	if err != nil {
		log.Print(err)
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

func loadNonce() *[24]byte {
	r := [24]byte{}
	bytes := make([]byte, 3)
	rand.Read(bytes)
	copy(r[:], string(bytes))
	return &r
}
