package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

// A SecureReader reads encrypted messages over
// its reader r and decrypts them with its key.
type SecureReader struct {
	key key
	r   io.Reader
}

// A SecureWriter encrypts messages with its key
// and writes them over its writer w.
type SecureWriter struct {
	key key
	w   io.Writer
}

// A secureConn wraps a connection with a secureReader and secureWriter
// to exchange encrypted messages
type secureConn struct {
	conn net.Conn
	*SecureReader
	*SecureWriter
}

// newSecureConn returns a secure connection that wraps the passed in conn
// with a SecurReader and SecureWriter built from a priv/pub key pair and
// the peer public key
func newSecureConn(conn net.Conn, priv, pub, peerPub *[32]byte) *secureConn {
	return &secureConn{
		conn,
		NewSecureReader(conn, priv, peerPub),
		NewSecureWriter(conn, priv, peerPub),
	}
}

// Close closes the connection.
func (s secureConn) Close() error {
	return s.conn.Close()
}

// NewSecureReader instantiates a new SecureReader.
func NewSecureReader(r io.Reader, priv, pub *[32]byte) *SecureReader {
	key := sharedKey(priv, pub)
	return &SecureReader{key, r}
}

// NewSecureWriter instantiates a new SecureWriter.
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) *SecureWriter {
	key := sharedKey(priv, pub)
	return &SecureWriter{key, w}
}

// Read wraps the underlying reader Read and provides decryption.
func (sr *SecureReader) Read(p []byte) (n int, err error) {
	// Read the encrypted message
	n, err = sr.r.Read(p)
	if err != nil {
		return n, err
	}
	p = p[:n]

	// decrypt the message
	msg, err := decrypt(p, sr.key)
	if err != nil {
		return n, err
	}
	n = copy(p, msg)
	return n, nil
}

// Write wraps the underlying writer Write and provides encryption.
func (sw *SecureWriter) Write(p []byte) (n int, err error) {
	c, err := encrypt(p, sw.key)
	if err != nil {
		return 0, err
	}
	return sw.w.Write(c)
}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	// generate the key pair for the client
	pub, priv, err := generateKeys()
	if err != nil {
		return nil, err
	}

	// connect to the server
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	// exchange public keys with the server
	conn.Write(pub[:])
	serverPub := &[32]byte{}
	conn.Read(serverPub[:])

	return newSecureConn(conn, priv, pub, serverPub), nil
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	// generate the server key pair
	pub, priv, err := generateKeys()
	if err != nil {
		return err
	}

	// wait for connections to serve
	// TODO in future, more complete versions, we need a mechanism to shut
	// down the server and gracioulsy shut down opened connections
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go handleServerConnection(conn, pub, priv)
	}
}

// handleServerConnection manages a client connection to the server.
func handleServerConnection(conn net.Conn, pub, priv *[32]byte) {
	defer conn.Close()
	// key exchange
	clientPub := &[32]byte{}
	conn.Read(clientPub[:])
	conn.Write(pub[:])

	// read client ciphertext
	c := make([]byte, maxCipherLen)
	n, err := conn.Read(c)
	if err != nil {
		fmt.Printf("Exiting connection on read: %v\n", err)
		return
	}
	c = c[:n]

	// decrypt ciphertext
	key := sharedKey(priv, clientPub)
	msg, err := decrypt(c, key)
	if err != nil {
		fmt.Printf("Exiting connectiond on decryption: %v\n", err)
		return
	}

	// encrypt and echo back the message
	c, err = encrypt(msg, key)
	if err != nil {
		fmt.Printf("Exiting connection on encryption: %v\n", err)
		return
	}

	n, err = conn.Write(c)
	if err != nil {
		fmt.Printf("Exiting connection on write: %v\n", err)
		return
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
		fmt.Printf("Server started listening on port %d\n", *port)
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
	defer conn.Close()
	msg := []byte(os.Args[2])
	if len(msg) > maxMsgLen {
		log.Fatalf("Message is too long, sorry. Max lenght is %v bytes,"+
			"yours is %v bytes\n", maxMsgLen, len(msg))
	}
	if _, err := conn.Write(msg); err != nil {
		log.Fatal(err)
	}
	buf := make([]byte, maxCipherLen)
	n, err := conn.Read(buf)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", buf[:n])
}
