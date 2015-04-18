package main

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"

	"golang.org/x/crypto/nacl/box"
)

// SecureReader reads a stream and decodes it
type SecureReader struct {
	R        io.Reader
	Prv, Pub *[32]byte
	br       *bytes.Reader
}

func (sr *SecureReader) Read(p []byte) (n int, err error) {
	var (
		nonce  [24]byte
		length int64
	)

	if err := binary.Read(sr.R, binary.LittleEndian, nonce[:]); err != nil {
		return 0, err
	}

	if err := binary.Read(sr.R, binary.LittleEndian, &length); err != nil {
		return 0, err
	}

	lr := io.LimitReader(sr.R, length)
	b, err := ioutil.ReadAll(lr)

	if err != nil {
		return 0, err
	}

	buf, ok := box.Open(nil, b, &nonce, sr.Pub, sr.Prv)

	if !ok {
		return 0, errors.New("decryption failed")
	}

	if sr.br == nil || sr.br.Len() == 0 {
		sr.br = bytes.NewReader(buf)
	}

	return sr.br.Read(p)
}

// SecureWriter generates a nonce, encodes data and writes the nonce, length of
// message to a stream
type SecureWriter struct {
	W        io.Writer
	Prv, Pub *[32]byte
	entropy  io.Reader
}

func (sw *SecureWriter) Write(p []byte) (int, error) {
	var nonce [24]byte
	sw.entropy.Read(nonce[:])

	buf := box.Seal(nil, p, &nonce, sw.Pub, sw.Prv)
	length := int64(len(buf))

	if err := binary.Write(sw.W, binary.LittleEndian, nonce[:]); err != nil {
		// No bytes from the original datasource have been written at this point
		return 0, err
	}

	if err := binary.Write(sw.W, binary.LittleEndian, length); err != nil {
		// No bytes from the original datasource have been written at this point
		return 0, err
	}

	if _, err := sw.W.Write(buf); err != nil {
		// We don't actually know how many bytes were written of the original source
		// as the encrypted data is larger
		return 0, err
	}

	return len(p), nil
}

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	return &SecureReader{r, priv, pub, nil}
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	return &SecureWriter{w, priv, pub, rand.Reader}
}

// SecureConn wraps net.Conn and encryptes/decryptes messages passed to it
type SecureConn struct {
	SecureReader
	SecureWriter
	conn net.Conn
}

// Close the connection
func (s *SecureConn) Close() error {
	return s.conn.Close()
}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	rpub := &[32]byte{}
	lpub, lprv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	binary.Read(conn, binary.LittleEndian, rpub)
	binary.Write(conn, binary.LittleEndian, lpub)

	return &SecureConn{
		SecureReader{conn, lprv, rpub, nil},
		SecureWriter{conn, lprv, rpub, rand.Reader},
		conn,
	}, nil
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	defer l.Close()

	rpub := &[32]byte{}
	lpub, lprv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			// TODO Handel errors
			return err
		}

		go func(conn net.Conn, lpub, lprv *[32]byte) {
			binary.Write(conn, binary.LittleEndian, lpub)
			binary.Read(conn, binary.LittleEndian, rpub)

			sconn := &SecureConn{
				SecureReader{conn, lprv, rpub, nil},
				SecureWriter{conn, lprv, rpub, rand.Reader},
				conn,
			}
			io.Copy(sconn, sconn)
		}(conn, lpub, lprv)
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
