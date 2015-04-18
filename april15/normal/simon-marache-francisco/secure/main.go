package main

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"golang.org/x/crypto/nacl/box"
	"io"
	"log"
	"net"
	"os"
)

// writeBinaryData is a helper used to chain multiple call to binary.Write
// without checking each time for the return value. If err != nil, this
// function will return err without doing any action
func writeBinaryData(w io.Writer, data interface{}, err error) error {
	if err != nil {
		return err
	}

	return binary.Write(w, binary.LittleEndian, data)
}

// readBinaryData is a helper used to chain multiple call to binary.Read
// without checking each time for the return value. If err != nil, this
// function will return err without doing any action
func readBinaryData(r io.Reader, data interface{}, err error) error {
	if err != nil {
		return err
	}

	return binary.Read(r, binary.LittleEndian, data)
}

// SecureReader can decrypt data encrypted through a
// SecureWriter provided that the correct keys are provided
type SecureReader struct {
	r             io.Reader
	peerPub       *[32]byte
	priv          *[32]byte
	remainingData *bytes.Buffer
}

// SecureWriter can encrypt data that will be decrypted
// through a SecureReader  provided that the correct
// keys are provided
type SecureWriter struct {
	w       io.Writer
	peerPub *[32]byte
	priv    *[32]byte
}

// Write will encrypt data and then Write it to the underlying
// io.Writer. the return value is the size of data, and not the
// amount of encrypted data Written.
func (sw SecureWriter) Write(data []byte) (int, error) {
	var boxed []byte
	var nonce [24]byte
	var err error
	var boxedLen int64

	// Generate random nonce. Risks of Collision are negligible
	// The nonce MUST be different for each call to box.Seal
	err = readBinaryData(rand.Reader, &nonce, err)
	if err != nil {
		return 0, err
	}

	boxed = box.Seal(boxed, data, &nonce, sw.peerPub, sw.priv)

	boxedLen = int64(len(boxed))

	// A message is, in order: the nonce, the length of the encrypted
	// data, and finally the encrypted data
	err = writeBinaryData(sw.w, &nonce, err)
	err = writeBinaryData(sw.w, &boxedLen, err)
	err = writeBinaryData(sw.w, boxed, err)

	if err != nil {
		return 0, err
	}
	return len(data), err
}

// Read will read encrypted data from the underlying io.Reader,
// and then decrypt it. the return value is the size copied to data,
// and not the amount of encrypted data Read.
func (sr SecureReader) Read(data []byte) (n int, err error) {

	// If the last call to Read didn't get the whole message, return
	// the part remaining. We can ignore the error since it will only be EOF
	if r, _ := sr.remainingData.Read(data); r > 0 {
		return r, nil
	}

	var boxedLen int64
	var nonce [24]byte

	err = readBinaryData(sr.r, &nonce, err)
	err = readBinaryData(sr.r, &boxedLen, err)
	if err != nil {
		return 0, err
	}

	boxed := make([]byte, boxedLen)

	err = readBinaryData(sr.r, &boxed, err)
	if err != nil {
		return 0, err
	}

	var out []byte
	out, isOk := box.Open(out, boxed, &nonce, sr.peerPub, sr.priv)
	if !isOk {
		return 0, errors.New("Couldn't decode input data")
	}

	// The tricky part: The caller want to read a 5 byte message, but we receive
	// a 200 byte message. The bytes that does not fit in data are saved in
	// the Buffer remainingData, and will be returned on a later call to Read
	// a new unit test has been created to check this behaviour: TestSmallReadPing
	copied := copy(data, out)
	if copied < len(out) {
		_, err := sr.remainingData.Write(out[copied:])
		if err != nil {
			return copied, err
		}
	}

	return copied, nil
}

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, peerPub *[32]byte) io.Reader {
	sec := SecureReader{
		priv:          priv,
		peerPub:       peerPub,
		r:             r,
		remainingData: bytes.NewBuffer([]byte{}),
	}

	return sec
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, peerPub *[32]byte) io.Writer {
	sec := SecureWriter{
		priv:    priv,
		peerPub: peerPub,
		w:       w,
	}

	return sec
}

// DialReadWriteCloser is a helper that implement the ReadWriteCloser
// interface for the return of the Dial() function. conn is the underlying
// socket.
type DialReadWriteCloser struct {
	sr   io.Reader
	sw   io.Writer
	conn io.ReadWriteCloser
}

// Read up to len(p) bytes from the secure channel
func (srw DialReadWriteCloser) Read(p []byte) (n int, err error) {
	return srw.sr.Read(p)
}

// Write p to the secure channel
func (srw DialReadWriteCloser) Write(p []byte) (n int, err error) {
	return srw.sw.Write(p)
}

// CLose the secure channel
func (srw DialReadWriteCloser) Close() (err error) {
	return srw.conn.Close()
}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	var peerPub [32]byte

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	err = writeBinaryData(conn, pub, err)
	err = readBinaryData(conn, &peerPub, err)

	if err != nil {
		return nil, err
	}

	a := DialReadWriteCloser{
		conn: conn,
		sr:   NewSecureReader(conn, priv, &peerPub),
		sw:   NewSecureWriter(conn, priv, &peerPub),
	}

	return a, nil
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	var peerPub [32]byte
	message := make([]byte, 2048)

	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go func(conn net.Conn) {
			defer conn.Close()

			err = writeBinaryData(conn, pub, err)
			err = readBinaryData(conn, &peerPub, err)
			if err != nil {
				fmt.Println("Error during handshake exchange: " + err.Error())
				return
			}

			sr := NewSecureReader(conn, priv, &peerPub)
			sw := NewSecureWriter(conn, priv, &peerPub)

			fmt.Println("Serve: starting Ping message loop")
			for err == nil {
				n, err := sr.Read(message)
				err = writeBinaryData(sw, message[:n], err)
				if err != nil {
					if err != io.EOF {
						fmt.Println("Error during message exchange: " + err.Error())
					}
					return
				}
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
