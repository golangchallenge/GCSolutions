package main

import (
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

// Maximum lenght of message whisch has to be encrypted
const MaxMsgLen = 32 * 1000

func secureRead(r io.Reader, p []byte, sharedKey *[32]byte) (n int, err error) {
	if len(p) > MaxMsgLen {
		p = p[:MaxMsgLen]
	}

	nonce := &[24]byte{}
	r.Read(nonce[:])
	//log.Printf("Readed nonce: %v\n", nonce)

	var msgLen uint16
	err = binary.Read(r, binary.LittleEndian, &msgLen)
	if err != nil {
		return
	}
	//log.Printf("Readed message length: %v\n", msgLen)

	b := make([]byte, msgLen)
	n, err = r.Read(b)
	if err != nil {
		return
	}
	b = b[:n]
	//log.Printf("Readed box: %v\n", b)

	o, ok := box.OpenAfterPrecomputation(nil, b, nonce, sharedKey)
	n = copy(p, o)
	//log.Printf("Reading: %s\n", p)
	if ok == false {
		err = ErrDecrypt
	}
	return
}

func secureWrite(w io.Writer, p []byte, sharedKey *[32]byte) (n int, err error) {
	if len(p) > MaxMsgLen {
		p = p[:MaxMsgLen]
	}
	//log.Printf("Writing: %s\n", p)

	nonce := &[24]byte{}
	rand.Read(nonce[:])

	//log.Printf("Writing nonce: %v\n", nonce)
	n, err = w.Write(nonce[:])
	if err != nil {
		return
	}

	o := box.SealAfterPrecomputation(nil, p, nonce, sharedKey)

	//log.Printf("Writing encrypted message length: %v\n", uint16(len(o)))
	err = binary.Write(w, binary.LittleEndian, uint16(len(o)))
	if err != nil {
		return
	}

	//log.Printf("Writing box: %v\n", o)
	i, err := w.Write(o)
	n = n + i
	if err != nil {
		return
	}
	n = len(p)
	return
}

type secureReader struct {
	sharedKey *[32]byte
	reader    io.Reader
}

// Error message when decrypt in reader fails
var ErrDecrypt = errors.New("secure: Decryption error")

func (s secureReader) Read(p []byte) (n int, err error) {
	n, err = secureRead(s.reader, p, s.sharedKey)
	return
}

// NewSecureReader instantiates a new secureReader
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	s := secureReader{sharedKey: &[32]byte{}, reader: r}
	box.Precompute(s.sharedKey, pub, priv)
	return s
}

type secureWriter struct {
	sharedKey *[32]byte
	writer    io.Writer
}

func (s secureWriter) Write(p []byte) (n int, err error) {
	n, err = secureWrite(s.writer, p, s.sharedKey)
	return
}

// NewSecureWriter instantiates a new secureWriter
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	s := secureWriter{sharedKey: &[32]byte{}, writer: w}
	box.Precompute(s.sharedKey, pub, priv)
	return s
}

type secureReadWriteCloser struct {
	sharedKey *[32]byte
	conn      io.ReadWriteCloser
}

// NewSecureReadWriteCloser instantiates a new secureWriter
func NewSecureReadWriteCloser(c io.ReadWriteCloser, priv, pub *[32]byte) io.ReadWriteCloser {
	s := secureReadWriteCloser{sharedKey: &[32]byte{}, conn: c}
	box.Precompute(s.sharedKey, pub, priv)
	return s
}

func (s secureReadWriteCloser) Read(p []byte) (n int, err error) {
	n, err = secureRead(s.conn, p, s.sharedKey)
	return
}

func (s secureReadWriteCloser) Write(p []byte) (n int, err error) {
	n, err = secureWrite(s.conn, p, s.sharedKey)
	return
}

func (s secureReadWriteCloser) Close() error {
	return s.conn.Close()
}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	cPub, cPriv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	sPub := &[32]byte{}
	_, err = conn.Read(sPub[:])
	if err != nil {
		return nil, err
	}

	_, err = conn.Write(cPub[:])
	if err != nil {
		return nil, err
	}

	s := NewSecureReadWriteCloser(conn, cPriv, sPub)
	return s, nil
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go func(conn net.Conn) {
			defer conn.Close()
			sPub, sPriv, err := box.GenerateKey(rand.Reader)
			if err != nil {
				log.Fatal(err)
				return
			}

			_, err = conn.Write(sPub[:])
			if err != nil {
				log.Fatal(err)
				return
			}

			cPub := &[32]byte{}
			_, err = conn.Read(cPub[:])
			if err != nil {
				log.Fatal(err)
				return
			}

			s := NewSecureReadWriteCloser(conn, sPriv, cPub)

			msg := make([]byte, MaxMsgLen)
			n, err := s.Read(msg)
			if err != nil {
				log.Fatal(err)
				return
			}
			msg = msg[:n]
			_, err = s.Write(msg)
			if err != nil {
				log.Fatal(err)
				return
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
