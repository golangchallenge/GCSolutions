package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"golang.org/x/crypto/nacl/box"
	"io"
	"log"
	"net"
	"os"
)

var nonce [24]byte

// SecureStream implements Reader Writer and Closer interface. Reader
// and Writer interface employs public key cryptography to allow
// encrypted message transfer.
type SecureStream struct {
	r         io.Reader
	w         io.Writer
	priv, pub *[32]byte
}

func (st SecureStream) Read(b []byte) (int, error) {
	encmsg := make([]byte, len(b))
	n, err := st.r.Read(encmsg)

	if err != nil {
		return n, err
	}

	tmp, _ := box.Open(nil, encmsg[:n], &nonce, st.pub, st.priv)
	copy(b, tmp)

	return len(tmp), nil
}

func (st SecureStream) Write(b []byte) (int, error) {
	// Generate nonce
	rand.Read(nonce[:])
	encmsg := box.Seal(nil, b, &nonce, st.pub, st.priv)
	return st.w.Write(encmsg)
}

func closer(c io.ReadWriteCloser) error {
	if err := c.Close(); err != nil {
		return err
	}
	return nil
}

// Close closes the reader and writer of SecureStream
func (st SecureStream) Close() error {
	// Since we assign same conn to both reader and writer closing
	// one should be sufficient.
	switch st.r.(type) {
	case nil:
		return nil
	case net.Conn:
		return closer(st.r.(io.ReadWriteCloser))
	}

	return nil
}

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	return SecureStream{r: r, priv: priv, pub: pub}
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	return SecureStream{w: w, priv: priv, pub: pub}
}

// NewSecureReadWriteCloser instantiates a new SecureReadWriteCloser
func NewSecureReadWriteCloser(rw io.ReadWriteCloser, priv, pub *[32]byte) io.ReadWriteCloser {
	return SecureStream{r: rw, w: rw, priv: priv, pub: pub}
}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	ownPub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	conn, cerr := net.Dial("tcp", addr)
	if cerr != nil {
		return nil, cerr
	}

	// Exchange the public key with peer
	if _, err = conn.Write(ownPub[:]); err != nil {
		return nil, err
	}

	// Lets read peers public key
	otherPub := [32]byte{}
	if _, err := conn.Read(otherPub[:]); err != nil {
		return nil, err
	}

	// Connection wrapped as SecureStream with public key of
	// server.
	return NewSecureReadWriteCloser(conn, priv, &otherPub), nil
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	ownPub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		// First finish the hand shake
		theirPub := [32]byte{}
		if _, err = conn.Read(theirPub[:]); err != nil {
			return err
		}

		if _, err = conn.Write(ownPub[:]); err != nil {
			return err
		}

		// Connection wrapped in SecureStream with public key
		// of Client.
		sstream := NewSecureReadWriteCloser(conn, priv, &theirPub)
		go func(c io.ReadWriteCloser) {
			defer c.Close()

			// Echo whatever read from client back to it
			io.Copy(c, c)
		}(sstream)
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
	if err != nil && err != io.EOF {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", buf[:n])
}
