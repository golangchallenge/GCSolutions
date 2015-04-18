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
	"sync"

	"golang.org/x/crypto/nacl/box"
)

// NaClReader is a read-and-decrypt wrapper of the underlying Reader
type NaClReader struct {
	Reader io.Reader // underlying reader
	skey   *[32]byte // secret decryption key
}

// NaClWriter is a encrypt-and-write wrapper of the underlying Writer
type NaClWriter struct {
	Writer io.Writer // underlying writer
	skey   *[32]byte // secret encryption key
}

// Read reads and decrypts the byte stream from the underlying Reader
func (r *NaClReader) Read(p []byte) (n int, err error) {
	buf := make([]byte, 1024*32)
	buflen, err := r.Reader.Read(buf)
	if buflen > 24 {
		// parse first 24 bytes as nounce
		nounce := new([24]byte)
		copy(nounce[:], buf[:24])

		// open the encrypted message
		opened, ok := box.OpenAfterPrecomputation(nil, buf[24:buflen], nounce, r.skey)
		if !ok {
			return 0, errors.New("unable to decrypt")
		}

		if len(p) < len(opened) {
			return 0, errors.New("read buffer exceeded")
		}

		n = copy(p, opened)
		return n, nil

	}

	if err != nil {
		return n, err
	}

	return 0, errors.New("unable to read enough data to decrypt")
}

// Write encrypts then write the byte stream to the underlying Writer
func (w *NaClWriter) Write(p []byte) (n int, err error) {
	nounce := new([24]byte)
	_, err = io.ReadFull(rand.Reader, nounce[:])
	if err != nil {
		return 0, errors.New("not able to obtain nonce")
	}
	sealed := box.SealAfterPrecomputation(nil, p, nounce, w.skey)
	if len(sealed) != len(p)+box.Overhead {
		return 0, errors.New("not able to encrypt all data")
	}

	// prepend nouce to sealed
	sealed = append(nounce[:], sealed...)
	return w.Writer.Write(sealed)
}

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	secretkey := new([32]byte)
	box.Precompute(secretkey, pub, priv)

	return &NaClReader{r, secretkey}
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	secretkey := new([32]byte)
	box.Precompute(secretkey, pub, priv)

	return &NaClWriter{w, secretkey}
}

// Conn is a secured connection using NaCl. It impelments
// the io.ReadWriteCloser interface.
type Conn struct {
	rawConn        net.Conn
	priv           *[32]byte
	pub            *[32]byte
	peerPublicKey  *[32]byte
	handshakeMutex sync.Mutex
	handshakeErr   error
}

// Read does handshake first and then read
func (c *Conn) Read(p []byte) (n int, err error) {
	// handles handshake and error
	err = c.Handshake()
	if err != nil {
		return
	}

	reader := NewSecureReader(c.rawConn, c.priv, c.peerPublicKey)
	return reader.Read(p)
}

// Write does handshake first and then write
func (c *Conn) Write(p []byte) (n int, err error) {
	err = c.Handshake()
	if err != nil {
		return
	}

	writer := NewSecureWriter(c.rawConn, c.priv, c.peerPublicKey)
	return writer.Write(p)
}

// Close closes the connection
func (c *Conn) Close() error {
	c.handshakeMutex.Lock()
	defer c.handshakeMutex.Unlock()
	if err := c.rawConn.Close(); err != nil {
		return err
	}
	return nil
}

// Handshake does the peering handshake.
// 1. Send peer the public key
// 2. obtain the peer public key
// use (peerPublicKey, priv) to seal/open messages
func (c *Conn) Handshake() error {
	c.handshakeMutex.Lock()
	defer c.handshakeMutex.Unlock()
	if err := c.handshakeErr; err != nil {
		return err
	}

	// presence of peerPublicKey signals the complete of handshake
	if c.peerPublicKey != nil {
		return nil
	}

	err := binary.Write(c.rawConn, binary.LittleEndian, c.pub)
	if err != nil {
		c.handshakeErr = err
		return err
	}

	c.peerPublicKey = new([32]byte)
	binary.Read(c.rawConn, binary.LittleEndian, c.peerPublicKey)
	if err != nil {
		c.handshakeErr = err
		return err
	}

	return nil
}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	rawConn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &Conn{rawConn: rawConn, priv: priv, pub: pub}, nil
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	defer l.Close()
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}

	for {
		rawConn, err := l.Accept()
		if err != nil {
			return err
		}
		conn := &Conn{rawConn: rawConn, pub: pub, priv: priv}
		go func(c *Conn) {
			defer c.Close()
			buf := make([]byte, 1024*32)
			n, err := c.Read(buf)
			if err != nil {
				fmt.Println(err)
			}
			got := string(buf[:n])
			log.Print(got)
			// echo the message
			n, err = c.Write([]byte(got))
			if err != nil {
				fmt.Println(err)
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
