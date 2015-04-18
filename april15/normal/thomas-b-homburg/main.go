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

type readWriteCloser struct {
	io.Reader
	io.Writer
	io.Closer
}

type secureReader struct {
	reader io.Reader
	priv   *[32]byte
	pub    *[32]byte
}

func (s *secureReader) Read(p []byte) (int, error) {

	// Read nonce
	var nonce [24]byte
	nonceSlice := nonce[:]
	s.reader.Read(nonceSlice)

	// Read raw encrypted message
	buf := make([]byte, 2048)
	n, err := s.reader.Read(buf)
	if nil != err {
		return n, err
	}
	buf = buf[:n]

	out, did := box.Open(nil, buf, &nonce, s.pub, s.priv)

	if did {
		n = copy(p, out)
		p = p[:n]
	}

	return n, nil
}

type secureWriter struct {
	writer io.Writer
	priv   *[32]byte
	pub    *[32]byte
}

func (w *secureWriter) Write(p []byte) (int, error) {

	// Generate and write nonce
	var nonce [24]byte
	nonceSlice := nonce[:]
	rand.Read(nonceSlice)
	w.writer.Write(nonceSlice)

	write := box.Seal(nil, p, &nonce, w.pub, w.priv)

	return w.writer.Write(write)
}

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	return &secureReader{r, priv, pub}
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	return &secureWriter{w, priv, pub}
}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	pub, priv, err := box.GenerateKey(rand.Reader)
	if nil != err {
		return nil, err
	}

	conn, err := net.Dial("tcp", addr)
	if nil != err {
		return nil, err
	}

	// Send client public key
	pubSlice := pub[:]
	_, err = conn.Write(pubSlice)
	if nil != err {
		return nil, err
	}

	// Get server public key
	var serverKey [32]byte
	serverKeySlice := serverKey[:]
	_, err = conn.Read(serverKeySlice)
	if nil != err {
		return nil, err
	}

	rwc := readWriteCloser{
		NewSecureReader(conn, priv, &serverKey),
		NewSecureWriter(conn, priv, &serverKey),
		conn,
	}
	return rwc, nil
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		go func(c io.ReadWriteCloser) {
			// Generate keys
			pub, priv, err := box.GenerateKey(rand.Reader)
			if nil != err {
				panic(err)
			}

			// Read client public key
			var clientKey [32]byte
			clientKeySlice := clientKey[:]
			_, err = c.Read(clientKeySlice)
			if nil != err {
				panic(err)
			}

			// Write server public key
			_, err = c.Write(pub[:])
			if err != nil {
				panic(err)
			}

			// Read from conn via secure reader
			p := make([]byte, 4096)
			r := NewSecureReader(c, priv, &clientKey)
			n, err := r.Read(p)

			if err == io.EOF {
				return
			} else if err != nil {
				panic(err)
			}

			w := NewSecureWriter(c, priv, &clientKey)

			// Echo message back via secure writer
			w.Write(p[:n])
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
