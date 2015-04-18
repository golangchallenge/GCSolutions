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

// messageHeader is the metadata sent along with encrypted message.
type messageHeader struct {
	Nonce [24]byte
	Size  uint16
}

// SecureReader is a wrapper around golang.org/x/crypto/nacl/box
// that implements io.Reader interface.
type SecureReader struct {
	r         io.Reader
	priv, pub *[32]byte
}

// NewSecureReader instantiates a new SecureReader.
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	return &SecureReader{r, priv, pub}
}

// Read reads a message, and decrypts it into p. The return value n
// is the number of bytes decrypted.
func (r *SecureReader) Read(p []byte) (n int, err error) {
	// Read the header
	var header messageHeader
	if err := binary.Read(r.r, binary.BigEndian, &header); err != nil {
		return 0, err
	}

	// Read the message and decrypt it
	emsg := make([]byte, header.Size)
	if err := binary.Read(r.r, binary.BigEndian, &emsg); err != nil {
		return 0, err
	}
	msg, ok := box.Open(nil, emsg, &header.Nonce, r.pub, r.priv)
	n = copy(p, msg)
	if !ok {
		err = errors.New("error decrypting message")
	}
	if n == 0 && err == nil {
		err = io.EOF
	}
	return
}

// SecureWriter is a wrapper around golang.org/x/crypto/nacl/box
// that implements io.Writer interface.
type SecureWriter struct {
	w         io.Writer
	priv, pub *[32]byte
}

// NewSecureWriter instantiates a new SecureWriter.
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	return &SecureWriter{w, priv, pub}
}

// Write encrypts the content of p and writes it to the underlying io.Writer
// along with metadata needed to decrypt the message.
// The return value n is the size of the plaintext message written.
func (w *SecureWriter) Write(p []byte) (n int, err error) {
	// Construct and write the header
	var header messageHeader
	if _, err := rand.Read(header.Nonce[:]); err != nil {
		return 0, err
	}
	header.Size = uint16(len(p) + box.Overhead)
	if err := binary.Write(w.w, binary.BigEndian, header); err != nil {
		return 0, err
	}

	// Encrypt and write the message
	emsg := box.Seal(nil, p, &header.Nonce, w.pub, w.priv)
	if err := binary.Write(w.w, binary.BigEndian, emsg); err != nil {
		return 0, err
	}
	return len(p), nil
}

// SecureConn is a connection that is capable of
// reading and writing encrypted messages.
type SecureConn struct {
	io.Reader
	io.Writer
	io.Closer
}

// NewSecureConn instantiates a new SecureConn.
func NewSecureConn(conn io.ReadWriteCloser, priv, pub *[32]byte) io.ReadWriteCloser {
	return &SecureConn{
		NewSecureReader(conn, priv, pub),
		NewSecureWriter(conn, priv, pub),
		conn,
	}
}

// exchangeKeys performs a key exchange with a remote peer
// and returns peer's public key after successfull exchange.
func exchangeKeys(conn io.ReadWriter, pub *[32]byte) (*[32]byte, error) {
	var ppub [32]byte
	if _, err := conn.Write(pub[:]); err != nil {
		return nil, err
	}
	if _, err := conn.Read(ppub[:]); err != nil {
		return nil, err
	}
	return &ppub, nil
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
	spub, err := exchangeKeys(conn, pub)
	if err != nil {
		return nil, err
	}
	return NewSecureConn(conn, priv, spub), nil
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}
	for {
		// Wait for a connection
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		// Handle the connection in a new goroutine
		go func(c net.Conn) {
			defer c.Close()
			cpub, err := exchangeKeys(c, pub)
			if err != nil {
				return
			}
			sc := NewSecureConn(c, priv, cpub)
			io.Copy(sc, sc)
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
