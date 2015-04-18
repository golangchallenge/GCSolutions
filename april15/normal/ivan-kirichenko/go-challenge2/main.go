package main

import (
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"golang.org/x/crypto/nacl/box"
)

// maxMessageLen defines maximum size of a message between client and server
// it is 32K plus overhead that NaCl adds
const maxMessageLen = 32768 + box.Overhead

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	key := [32]byte{}
	box.Precompute(&key, pub, priv)

	return &secureReader{r: r, key: &key}
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	key := [32]byte{}
	box.Precompute(&key, pub, priv)

	return &secureWriter{w: w, key: &key}
}

type secureWriter struct {
	w   io.Writer
	key *[32]byte
}

// Write ciphers incoming data `p` and writes it
func (w *secureWriter) Write(p []byte) (n int, err error) {
	nonce := [24]byte{}

	_, err = rand.Read(nonce[:])
	if err != nil {
		return
	}

	out := box.SealAfterPrecomputation(nonce[:], p, &nonce, w.key)

	return w.w.Write(out)
}

type secureReader struct {
	r io.Reader

	key      *[32]byte
	dataBuf  [maxMessageLen]byte
	nonceBuf [24]byte
	m        sync.Mutex
}

// Read reads secured data, deciphers it and populates incoming
// buffer p
// The method is thread-safe
func (r *secureReader) Read(p []byte) (n int, err error) {
	r.m.Lock()
	defer r.m.Unlock()

	// read nonce first
	if n, err = r.r.Read(r.nonceBuf[:]); err != nil {
		return
	} else if n != 24 {
		err = errors.New("could not read nonce")
		return
	}

	// read actual message and decipher it
	if n, err = r.r.Read(r.dataBuf[:]); err != nil {
		return
	}

	out, ok := box.OpenAfterPrecomputation([]byte{}, r.dataBuf[:n], &r.nonceBuf, r.key)
	if !ok {
		err = errors.New("could not decrypt incoming message")
		return
	}

	n = copy(p, out)
	if n < len(out) {
		err = fmt.Errorf("buffer is too small. Need to allocate %d bytes, got %d", len(out), n)
	}
	return
}

// secureConn is a wrapper around connection that can close connection
// implements io.ReadWriteCloser
type secureConn struct {
	w io.Writer
	r io.Reader

	c net.Conn
}

func newSecureConn(conn net.Conn, priv, pub *[32]byte) io.ReadWriteCloser {
	secReader := NewSecureReader(conn, priv, pub)
	secWriter := NewSecureWriter(conn, priv, pub)

	return &secureConn{r: secReader, w: secWriter, c: conn}
}

func (c *secureConn) Read(p []byte) (n int, err error) {
	return c.r.Read(p)
}

func (c *secureConn) Write(p []byte) (n int, err error) {
	return c.w.Write(p)
}

func (c *secureConn) Close() error {
	return c.c.Close()
}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (secConn io.ReadWriteCloser, err error) {
	// generate keys
	var pub, priv *[32]byte
	pub, priv, err = box.GenerateKey(rand.Reader)
	if err != nil {
		return
	}

	// open TCP connection
	var conn net.Conn
	conn, err = net.Dial("tcp", addr)
	if err != nil {
		return

	}

	// send own public key
	if _, err = conn.Write(pub[:]); err != nil {
		err = fmt.Errorf("could not send a public key: %s", err.Error())
		return
	}

	// receive server's public key
	var serverPub [32]byte
	if _, err = conn.Read(serverPub[:]); err != nil {
		err = fmt.Errorf("could not receive server public key: %s", err.Error())
		return
	}

	// create secure wrapper around connection
	secConn = newSecureConn(conn, priv, &serverPub)

	return
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) (err error) {
	// generate keys
	var pub, priv *[32]byte
	pub, priv, err = box.GenerateKey(rand.Reader)
	if err != nil {
		return
	}

	// buffer for incoming messages. 24 is nonce size
	buf := make([]byte, maxMessageLen+24)

	for {
		// accept connection
		conn, err := l.Accept()

		if err != nil {
			log.Println("could not accept connection: ", err.Error())
			time.Sleep(time.Second)
			continue
		}

		// read client's public key
		// not checking number of bytes read and written here
		// because in such situation decryption will fail anyway
		var clientPub [32]byte
		if _, err = conn.Read(clientPub[:]); err != nil {
			log.Println("could not receive client public key: ", err.Error())
			conn.Close()
			continue
		}

		// send server's public key to client
		if _, err = conn.Write(pub[:]); err != nil {
			log.Println("could not send public key to client: ", err.Error())
			conn.Close()
			continue
		}

		// create secure wrapper, read data from client and send it back
		secConn := newSecureConn(conn, priv, &clientPub)

		if n, err := secConn.Read(buf); err == nil {
			_, err = secConn.Write(buf[:n])
		}

		if err != nil {
			log.Println(err.Error())
		}

		secConn.Close()
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
