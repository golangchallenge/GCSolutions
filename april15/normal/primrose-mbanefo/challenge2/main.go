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

type secureReadWriteCloser struct {
	in      io.Reader
	out     io.Writer
	priv    *[32]byte
	peerPub *[32]byte
}

func (s secureReadWriteCloser) Read(p []byte) (n int, err error) {
	if s.in != nil {
		// read incoming data
		buf := make([]byte, 1024)
		n, err = s.in.Read(buf)
		if err != nil {
			return
		}

		// if there were no errors and
		// if there is data to decrypt
		if n > 0 {
			// decrypt the data
			buf, b := box.Open(nil, buf[:n], &nonce, s.peerPub, s.priv)
			// check for any errors
			if b {
				// use the decrypted data as the result of read
				n = copy(p, buf[:len(buf)])
			}
		}
	}
	return
}

func (s secureReadWriteCloser) Write(data []byte) (n int, err error) {
	if s.out != nil {
		// get a nonce
		// crypto rand seems capable of generating a 24byte prime number
		nonceBigint, err := rand.Prime(rand.Reader, 24*8)
		if err != nil {
			return n, err
		}

		copy(nonce[:], nonceBigint.Bytes())

		// send the encrypted data
		return s.out.Write(box.Seal(nil, data, &nonce, s.peerPub, s.priv))
	}
	return
}

func (s secureReadWriteCloser) Close() error {
	return nil
}

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	return secureReadWriteCloser{r, nil, priv, pub}
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	return secureReadWriteCloser{nil, w, priv, pub}
}

// NewSecureReadWriter instantiates a new SecureReaderWriter
func NewSecureReadWriter(r io.Reader, w io.Writer, priv, pub *[32]byte) io.ReadWriteCloser {
	return secureReadWriteCloser{r, w, priv, pub}
}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	//connect to server
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	// generate private/public key pair
	myPub, myPriv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	// Handshake: send keys to server
	n, err := conn.Write(myPub[:])
	if err != nil || n != len(myPub) {
		return nil, err
	}

	// Handshake: Get server's public key
	var peerPub [32]byte
	_, err = conn.Read(peerPub[:])
	if err != nil {
		return nil, err
	}

	// send over secure connection for future messages
	return NewSecureReadWriter(conn, conn, myPriv, &peerPub), err
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) (err error) {
	//accept connection
	conn, err := l.Accept()
	if err != nil {
		return
	}

	// generate private/public key pair
	myPub, myPriv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return
	}

	// Handshake: Get clients public key
	var peerPub [32]byte
	n, err := conn.Read(peerPub[:])
	if err != nil || n != len(peerPub) {
		// Stopping here would break the tests. Alert user instead
		fmt.Println("The received public key looks fairly dodgy")
	}

	// Handshake: send my public key
	n, err = conn.Write(myPub[:])
	if err != nil || n != len(myPub) {
		return err
	}

	//assume every other message is encrypted and the only action is to echo it
	secureConn := NewSecureReadWriter(conn, conn, myPriv, &peerPub)
	msg := make([]byte, 1024)
	for {
		n, err = secureConn.Read(msg)
		if err != nil {
			break
		}

		secureConn.Write(msg[:n])
	}

	return err
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
