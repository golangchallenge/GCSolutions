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

// SecureReader decrypts messages encrypted using the NaCl asymmetric
// cryptosystem.
type SecureReader struct {
	io.Reader
	priv, pub *[32]byte
}

// NewSecureReader instantiates a new SecureReader that will take data from the given
// io.Reader and utilize the given privte and public keys for decryption.
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	return SecureReader{
		Reader: r,
		priv:   priv,
		pub:    pub,
	}
}

// Read implements the io.Reader interface on secureReader. The length
// returned is the length of the read decrypted message.
func (sr SecureReader) Read(p []byte) (n int, err error) {
	encrypted := make([]byte, 32000)

	n, err = sr.Reader.Read(encrypted)
	if err != nil {
		return 0, err
	}

	// The encrypted message has the 24 byte nonce prepended
	nonce := new([24]byte)
	copy(nonce[:], encrypted[0:24])
	encrypted = encrypted[24:n]

	decrypted, ok := box.Open(nil, encrypted, nonce, sr.pub, sr.priv)
	if !ok {
		return 0, errors.New("unable to decrypt message")
	}

	copy(p, decrypted)

	return len(decrypted), nil
}

// SecureWriter writes encrypted messages using the NaCl asymmetric cryptosystem.
type SecureWriter struct {
	io.Writer
	priv, pub *[32]byte
}

// NewSecureWriter instantiates a new SecureWriter that will encrypt messages using the
// given private and public keys and write it to the given io.Writer.
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	return SecureWriter{
		Writer: w,
		priv:   priv,
		pub:    pub,
	}
}

// Write implements the io.Writer interface on secureWriter.
func (sw SecureWriter) Write(p []byte) (n int, err error) {
	nonce := new([24]byte)
	n, err = rand.Read(nonce[:])
	if n < 24 {
		return 0, errors.New("secureWriter: unable to get 24 bytes of randomnness for nonce")
	} else if err != nil {
		return 0, err
	}

	// We want the nonce prepended to the encrypted message
	out := make([]byte, 24, 24+len(p)+box.Overhead)
	copy(out[0:24], nonce[:])

	return sw.Writer.Write(box.Seal(out, p, nonce, sw.pub, sw.priv))
}

// SecureConn encrypts and decrypts messages using the NaCl asymmetric cryptosystem
type SecureConn struct {
	io.Writer
	io.Reader
	io.Closer
}

// NewSecureConn instantiates a new SecureConn utilizing the given private and
// public key
func NewSecureConn(priv, pub *[32]byte, c net.Conn) io.ReadWriteCloser {
	return SecureConn{
		Writer: NewSecureWriter(c, priv, pub),
		Reader: NewSecureReader(c, priv, pub),
		Closer: c,
	}
}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	// Perform the handshake to exchange public keys. Server goes first.
	serverPub := new([32]byte)
	_, err = io.ReadFull(conn, serverPub[:])
	if err != nil {
		return nil, err
	}
	_, err = conn.Write(pub[:])
	if err != nil {
		return nil, err
	}

	return NewSecureConn(priv, serverPub, conn), nil
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		// Spawn a goroutine to handle the conn and securely echo back the message
		go func(c net.Conn) {
			_, err := c.Write(pub[:])
			if err != nil {
				log.Println("Error writing key: ", err)
				c.Close()
				return
			}

			clientPub := new([32]byte)
			_, err = io.ReadFull(c, clientPub[:])
			if err != nil {
				log.Println("Error reading key: ", err)
				c.Close()
				return
			}

			// Create the secure connection now that the handshake is done
			sc := NewSecureConn(priv, clientPub, c)
			defer sc.Close()

			message := make([]byte, 32000)
			n, err := sc.Read(message)
			if err != nil {
				log.Println("Error reading message: ", err)
				return
			}
			message = message[:n]

			// We don't know how big the encrypted message will be, so ignore n
			_, err = sc.Write(message)
			if err != nil {
				log.Println("Error writing message: ", err)
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
