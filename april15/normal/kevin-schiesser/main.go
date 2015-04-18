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

	"golang.org/x/crypto/nacl/box"
)

const maxMsgLen = 32 * 1024

// randomNonce fills the 24 byte slice with random bytes
// and returns a pointer to a newly allocated 24 byte array of the same bytes.
func randomNonce(buf []byte) (*[24]byte, error) {
	var nonce [24]byte
	if n, err := rand.Read(buf); n != 24 {
		return nil, fmt.Errorf("nonce length error %d != 24", n)
	} else if err != nil {
		return nil, err
	}
	for idx, val := range buf {
		nonce[idx] = val
	}
	return &nonce, nil
}

// NewSecureReader instantiates a new SecureReader.
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	return &SecureReader{
		reader:         r,
		peersPublicKey: pub,
		privateKey:     priv,
	}
}

// SecureReader decrypts NaCl encrypted messages.
// Messages are of the format [cipherText, nonce].
type SecureReader struct {
	reader         io.Reader
	peersPublicKey *[32]byte
	privateKey     *[32]byte
}

// Read NaCl encrypted messages for the format [cipherText, nonce].
func (sr *SecureReader) Read(buf []byte) (n int, err error) {
	var out []byte
	var nonce [24]byte
	message := make([]byte, maxMsgLen)
	n, err = sr.reader.Read(message)
	if err != nil {
		return n, err
	}
	for idx, val := range message[:24] {
		nonce[idx] = val
	}
	out, ok := box.Open(out, message[24:n], &nonce, sr.peersPublicKey, sr.privateKey)
	if !ok {
		return 0, errors.New("failed to decrypt cipher text")
	}
	n = copy(buf, out)
	return n, nil
}

// NewSecureWriter instantiates a new SecureWriter.
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	return &SecureWriter{
		writer:         w,
		peersPublicKey: pub,
		privateKey:     priv,
	}
}

// SecureWriter writes NaCl encripted messages.
// Messages are of the format [cipherText, nonce].
type SecureWriter struct {
	writer         io.Writer
	peersPublicKey *[32]byte
	privateKey     *[32]byte
}

// Write NaCl encrypted messages for the format [cipherText, nonce].
// Write increments the nonce after every call.
func (sw *SecureWriter) Write(plainText []byte) (n int, err error) {
	if len(plainText) > maxMsgLen {
		return 0, fmt.Errorf("message too long %d > %d bytes", len(plainText), maxMsgLen)
	}
	var out []byte
	nonce := make([]byte, 24)
	nonceptr, err := randomNonce(nonce)
	if err != nil {
		return 0, err
	}
	out = box.Seal(out, plainText, nonceptr, sw.peersPublicKey, sw.privateKey)
	message := append(nonce, out...)
	n, err = sw.writer.Write(message)
	if err != nil {
		return n, err
	}
	return len(message), nil
}

// NewConnWrapper instantiates a new ConnWrapper.
func NewConnWrapper(r io.Reader, w io.Writer, c io.Closer) io.ReadWriteCloser {
	return &ConnWrapper{
		r: r,
		w: w,
		c: c,
	}
}

// ConnWrapper wraps a network connection with
// an instance of a SecureReader and SecureWriter.
// Closer is the network connection default.
type ConnWrapper struct {
	r io.Reader
	w io.Writer
	c io.Closer
}

// Read from connection with SecureReader implementation.
func (cw *ConnWrapper) Read(b []byte) (n int, err error) {
	n, err = cw.r.Read(b)
	return n, err
}

// Read to connection with SecureWriter implementation.
func (cw *ConnWrapper) Write(b []byte) (n int, err error) {
	n, err = cw.w.Write(b)
	return n, err
}

// Close the connection using net.Conn default method.
func (cw *ConnWrapper) Close() error {
	err := cw.c.Close()
	return err
}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	sconn, err := handshake(pub, priv, conn)
	if err != nil {
		return nil, err
	}
	return sconn, nil
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	defer l.Close()
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}
	for {
		conn, err := l.Accept()
		// Not accepting connections, stop serving
		if err != nil {
			return err
		}
		// good connection, shake hands, echo and close
		go func(conn io.ReadWriteCloser) {
			defer conn.Close()
			sconn, err := handshake(pub, priv, conn)
			// bad handshake, return and do not echo
			if err != nil {
				return
			}
			io.Copy(sconn, sconn)
		}(conn)
	}
}

// handshake implements the client/server key exchange
// and returns a secure io.ReadWriteCloser if successful
// else it returns an error.
// Reads and writes are preformed concurrently, allowing
// both the client and server to call the same function
// without knowing who will read/write first.
func handshake(pub, priv *[32]byte, conn io.ReadWriteCloser) (io.ReadWriteCloser, error) {
	readDone := make(chan error)
	writeDone := make(chan error)
	peersPublicKey := make([]byte, 32)
	ppkArray := [32]byte{}

	go func(readDone chan<- error) {
		if n, err := conn.Read(peersPublicKey); n != 32 {
			readDone <- fmt.Errorf("key length error %d != 32\n", n)
			return
		} else if err != nil {
			readDone <- err
			return
		}
		for idx, val := range peersPublicKey {
			ppkArray[idx] = val
		}
		readDone <- nil
	}(readDone)

	go func(writeDone chan<- error) {
		if _, err := conn.Write(pub[:]); err != nil {
			writeDone <- err
			return
		}
		writeDone <- nil
	}(writeDone)

	// wait for read to finish
	err := <-readDone
	if err != nil {
		return nil, err
	}

	// wait for write to finish
	err = <-writeDone
	if err != nil {
		return nil, err
	}

	sr := NewSecureReader(conn, priv, &ppkArray)
	sw := NewSecureWriter(conn, priv, &ppkArray)
	return NewConnWrapper(sr, sw, conn), nil
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
