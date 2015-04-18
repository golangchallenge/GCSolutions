package main

import (
	"bytes"
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

// SecureWriter struct provides a secure way to write data.
// SecureWriter imlements the io.Writer interface.
type SecureWriter struct {
	pubKey  *[32]byte
	privKey *[32]byte
	writer  io.Writer
}

//SecureReader struct provides a secure way to read data.
//SecureReader imlements the io.Reader interface.
type SecureReader struct {
	pubKey  *[32]byte
	privKey *[32]byte
	reader  io.Reader
}

//ConnRWC implements the io.ReadWriteCloser interface.
//It is used to wrap a secure reader and writer around a normal net.Conn.
type ConnRWC struct {
	io.Writer
	io.Reader
	connection net.Conn
}

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {

	sr := SecureReader{pubKey: pub, privKey: priv, reader: r}

	return sr
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {

	sw := SecureWriter{pubKey: pub, privKey: priv, writer: w}

	return sw
}

func (w SecureWriter) Write(p []byte) (int, error) {

	nonce := new([24]byte)

	//Just create random nonces every time ( as stated in NaCl documentation
	//randomness of 24 bytes provided enough entropy )
	_, err := io.ReadFull(rand.Reader, nonce[:])

	if err != nil {
		return 0, err
	}

	enc := box.Seal(nil, p, nonce, w.pubKey, w.privKey)

	m := new(bytes.Buffer)

	err = binary.Write(m, binary.LittleEndian, nonce)
	if err != nil {
		return 0, fmt.Errorf("Writing failed: %s", err)
	}

	err = binary.Write(m, binary.LittleEndian, enc)
	if err != nil {
		return 0, fmt.Errorf("Writing failed: %s", err)
	}

	return w.writer.Write(m.Bytes())

}

func (r SecureReader) Read(p []byte) (int, error) {

	b := new([24]byte)

	err := binary.Read(r.reader, binary.LittleEndian, b)
	if err != nil {
		return 0, fmt.Errorf("Reading failed: %s", err)
	}

	buf := make([]byte, 1024)

	rb, err := r.reader.Read(buf)

	buf = buf[:rb]

	res, valid := box.Open(nil, buf, b, r.pubKey, r.privKey)

	if !valid {
		return 0, errors.New("Could not decrypt received message")
	}

	d := copy(p, res)

	return d, nil

}

//Close is used to close the connection associated with the current ConnRWC.
func (c ConnRWC) Close() error {

	return c.connection.Close()

}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {

	//perform simple tcp connection
	conn, err := net.Dial("tcp", addr)

	if err != nil {
		return nil, err
	}

	pub, priv, err := box.GenerateKey(rand.Reader)

	buf := new(bytes.Buffer)

	err = binary.Write(buf, binary.LittleEndian, pub)
	if err != nil {
		return nil, fmt.Errorf("Handshake failed: %s", err)
	}

	//Send our public key to the server
	conn.Write(buf.Bytes())

	serverKey := new([32]byte)

	//Receive server's public key
	err = binary.Read(conn, binary.LittleEndian, serverKey)
	if err != nil {
		return nil, fmt.Errorf("Handshake failed: %s", err)
	}

	//create upgraded encrypted wrapper for the connection
	rwc := &ConnRWC{}

	rwc.Reader = NewSecureReader(conn, priv, serverKey)
	rwc.Writer = NewSecureWriter(conn, priv, serverKey)
	rwc.connection = conn

	return rwc, nil

}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {

	pub, priv, err := box.GenerateKey(rand.Reader)

	if err != nil {
		panic(err)
	}

	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		// Handle connections in a new goroutine.
		go func(c net.Conn) {

			clientKey := new([32]byte)

			err = binary.Read(c, binary.LittleEndian, clientKey)
			if err != nil {
				log.Println("binary.Read 2 failed:", err)
			}

			buf := new(bytes.Buffer)

			err = binary.Write(buf, binary.LittleEndian, pub)
			if err != nil {
				log.Println("binary.Write 2 failed:", err)
			}

			c.Write(buf.Bytes())

			r := NewSecureReader(c, priv, clientKey)
			w := NewSecureWriter(c, priv, clientKey)

			io.Copy(w, r)

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
	if err != nil && err != io.EOF {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", buf[:n])
}
