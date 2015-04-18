package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	"golang.org/x/crypto/nacl/box"
)

type SecureReader struct {
	io.Reader
	priv, pub *[32]byte
	closed    bool
}

// Read reads secure bytes from s and decrypts it into b
func (s *SecureReader) Read(b []byte) (int, error) {
	if s.closed {
		return 0, io.EOF
	}

	msg, err := getMsg(s)
	if err != nil {
		return 0, err
	}
	bb, ok := box.Open(nil, msg.body, msg.nonce, s.pub, s.priv)
	if !ok {
		return 0, fmt.Errorf("error decrypting message: %v ", msg)
	}
	n := copy(b, bb)
	return n, nil
}

type SecureWriter struct {
	io.Writer
	priv, pub *[32]byte
	closed    bool
}

// Write reads from b and writes it securely into s
func (s *SecureWriter) Write(b []byte) (int, error) {
	if s.closed {
		return 0, io.EOF
	}

	if len(b) == 0 {
		return 0, fmt.Errorf("No bytes to write")
	}

	nonce := newNonce()

	o := box.Seal(nil, b, nonce, s.pub, s.priv)
	var length byte = byte(len(o))

	if length == 0 {
		return 0, fmt.Errorf("Error writing secure message with nonce: %v", nonce)
	}

	var out []byte

	// nonce
	out = append(out, nonce[:]...)
	// length
	out = append(out, length)
	// encrypted bytes
	out = append(out, o...)
	return s.Writer.Write(out)
}

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, pub *[32]byte) *SecureReader {
	return &SecureReader{r, pub, priv, false}
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) *SecureWriter {
	return &SecureWriter{w, pub, priv, false}
}

// SecureReadWriteCloser embeds SecureReader and SecureWriter for secure reading and writing respectively
type SecureReadWriteCloser struct {
	*SecureWriter
	*SecureReader
}

// Close marks s as closed. Further reads or writes after a call to Close returns io.EOF
func (s *SecureReadWriteCloser) Close() error {
	s.SecureReader.closed = true
	s.SecureWriter.closed = true
	return nil
}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	var pub [32]byte
	n, err := conn.Read(pub[:])
	if n == 0 || err != nil {
		conn.Close()
		return nil, fmt.Errorf("Error retrieving public key from server, handshake failed")
	}
	priv, myPub, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("Handshake failed")
	}

	_, err = conn.Write(myPub[:])

	if err != nil {
		return nil, fmt.Errorf("Handshake failed")
	}

	return &SecureReadWriteCloser{
		NewSecureWriter(conn, priv, &pub),
		NewSecureReader(conn, priv, &pub),
	}, nil
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	for {
		conn, err := l.Accept()
		if err != nil {
			continue
			continue
		}
		go handleAndEcho(conn)
	}
}

// handleAndEcho handles a client connection to the server and echos the message body back to the client
func handleAndEcho(conn net.Conn) {
	priv, myPub, err := box.GenerateKey(rand.Reader)
	if err != nil {
		log.Fatal("Cannot Generate Key")
	}
	_, err = conn.Write(myPub[:])
	if err != nil {
		log.Fatal(err)
	}

	var pub [32]byte
	_, err = conn.Read(pub[:])

	if err != nil {
		log.Fatal("Handshake Failed")
	}

	s := &SecureReadWriteCloser{
		NewSecureWriter(conn, priv, &pub),
		NewSecureReader(conn, priv, &pub),
	}

	var buf [2048]byte

	n, err := s.Read(buf[:])
	if err != nil {
		log.Fatal(err)
	}
	_, err = s.Write(buf[:n])
	if err != nil {
		log.Fatal(err)
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

// msg is a representation of encrypted message
type msg struct {
	body  []byte
	nonce *[24]byte
}

// getMsg reads a new msg from s
func getMsg(s *SecureReader) (msg, error) {
	var nonce [24]byte
	// read nonce
	// block until there's something to read
	var err error
	for {
		_, err = s.Reader.Read(nonce[:])
		if err != io.EOF {
			break
		}
		time.Sleep(time.Millisecond * 100)
	}
	if err != nil {
		return msg{}, err
	}
	var l [1]byte
	_, err = s.Reader.Read(l[:])
	if err != nil {
		return msg{}, err
	}
	var length byte = l[0]
	body := make([]byte, length)
	_, err = s.Reader.Read(body)
	if err != nil {
		return msg{}, err
	}
	return msg{body, &nonce}, nil
}

// randBytes generates random bytes of length l
func randBytes(length int) []byte {
	var b = make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal("cannot generate random bytes :", err)
	}
	return b
}

// newNonce creates a new nonce to be used for message encryption
func newNonce() *[24]byte {
	var nonce [24]byte
	copy(nonce[:], randBytes(24))
	return &nonce
}
