package main

import (
	"crypto/rand"
	"encoding/binary"
	"golang.org/x/crypto/nacl/box"

	"flag"
	"io"
	"net"
	"os"

	"errors"
	"fmt"
	"log"
)

const (
	// MaxLength is the unencrypted max length
	MaxLength = 65536
	// NonceBytes is the bytes used for the nonce
	NonceBytes = 24
	// LengthBytes is the bytes used for the length
	LengthBytes = 4
	// MaxEncryptedLength is the encrypted packet max length
	MaxEncryptedLength = MaxLength + NonceBytes + LengthBytes + box.Overhead
)

// SecureReadWriteCloser is a NaCl powered crypto io.ReadWriteCloser
type SecureReadWriteCloser struct {
	secret *[32]byte
	reader io.Reader
	writer io.Writer
	closer io.Closer
}

// Error types for SecureReadWriteCloser
var (
	ErrOutputTooLong = errors.New("Output too long")
	ErrInputTooLong  = errors.New("Input too long")
	ErrSliceTooSmall = errors.New("Slice too small for packet")
	ErrRandom        = errors.New("Unable to fetch random data")
	ErrVerification  = errors.New("Unable to verify message")
)

// Reads an entire encrypted packet, and handles decryption. This blocks until
// the entire package have been read. Do note that any error apart from
// ErrVerification is unrecoverable, as the position in the stream would not
// be sane afterwards.
func (srwc SecureReadWriteCloser) Read(p []byte) (int, error) {
	// The structure of a packet for reference is:
	//      [ length uint32 | nonce [24]byte | enc_message []byte ]

	// Calling us with len(p) == 0 makes no sense
	if len(p) == 0 {
		return 0, nil
	}

	// We'd like to have the length
	var length [LengthBytes]byte
	if _, err := io.ReadFull(srwc.reader, length[:]); err != nil {
		return 0, err
	}

	// The length includes both the length and nonce bytes themselves, so
	// subtract that
	parsedLength := binary.BigEndian.Uint32(length[:]) - NonceBytes - LengthBytes
	if parsedLength > MaxLength {
		return 0, ErrInputTooLong
	}

	if int(parsedLength)-box.Overhead > len(p) {
		return 0, ErrSliceTooSmall
	}

	// ... And then we read a nonce
	var nonce [NonceBytes]byte
	if _, err := io.ReadFull(srwc.reader, nonce[:]); err != nil {
		return 0, err
	}

	// Fetch the message
	msg := make([]byte, parsedLength)
	if _, err := io.ReadFull(srwc.reader, msg); err != nil {
		return 0, err
	}

	// Decryyyyypt!
	out, verified := box.OpenAfterPrecomputation(nil, msg, &nonce, srwc.secret)
	if !verified {
		return 0, ErrVerification
	}

	copy(p, out)
	return len(out), nil
}

// Write handles encryption and sends the data off with nonce and length in the
// header. Do note that any error is unrecoverable if data has been sent, as
// the recipient will be unable to understand the stream.
func (srwc SecureReadWriteCloser) Write(p []byte) (int, error) {
	// The structure of a packet for reference is:
	//      [ length uint32 | nonce [24]byte | enc_message []byte ]

	// Test if the input is too long
	if len(p) > MaxLength {
		return 0, ErrOutputTooLong
	}

	length := uint32(len(p) + NonceBytes + LengthBytes + box.Overhead)

	// Prepare the nonce
	// Using the same nonce twice reduces the difficulty of "guessing" the key,
	// so that's a big no-no, but while random numbers have a risk for collision,
	// NaCL docs (http://nacl.cr.yp.to/stream.html) state that, due to the large
	// nonce size (24 bytes), that randomly generated nonces have negligible risk
	// of collision.
	nonce := new([NonceBytes]byte)
	if _, err := rand.Read(nonce[:]); err != nil {
		return 0, ErrRandom
	}

	// Make a slice with capacity for the entire thing, and the length set to the
	// length/nonce we'll be copying in
	message := make([]byte, LengthBytes+NonceBytes, length)

	binary.BigEndian.PutUint32(message[0:LengthBytes], length)
	copy(message[LengthBytes:LengthBytes+NonceBytes], nonce[:])

	// Encrypt!
	ret := box.SealAfterPrecomputation(message, p, nonce, srwc.secret)

	// Write all the things! We just use this as the return value.
	return srwc.writer.Write(ret)
}

// Close calls on the provided closer
func (srwc SecureReadWriteCloser) Close() error {
	return srwc.closer.Close()
}

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	sharedKey := new([32]byte)
	box.Precompute(sharedKey, pub, priv)

	return SecureReadWriteCloser{
		reader: r,
		secret: sharedKey,
	}
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	sharedKey := new([32]byte)
	box.Precompute(sharedKey, pub, priv)

	return SecureReadWriteCloser{
		writer: w,
		secret: sharedKey,
	}
}

// NewSecureReadWriteCloser instantiates a new SecureReadWriteCloser
func NewSecureReadWriteCloser(rwc io.ReadWriteCloser, priv, pub *[32]byte) io.ReadWriteCloser {
	sharedKey := new([32]byte)
	box.Precompute(sharedKey, pub, priv)

	return SecureReadWriteCloser{
		reader: rwc,
		writer: rwc,
		closer: rwc,
		secret: sharedKey,
	}
}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	publicKey, privateKey, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	// We read the remote public key
	var remotePubKey [32]byte
	if _, err = io.ReadFull(conn, remotePubKey[:]); err != nil {
		return nil, err
	}

	// And finish off by sending ours
	if _, err = conn.Write(publicKey[:]); err != nil {
		return nil, err
	}
	return NewSecureReadWriteCloser(conn, privateKey, &remotePubKey), nil
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	publicKey, privateKey, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		go func(c net.Conn) {
			// This is our connection handler. In case someone closes the
			// connection, or stops sending, we don't really care, and simply
			// close things nicely on our end as well.
			defer c.Close()

			// We send out our public key.
			if _, err := c.Write(publicKey[:]); err != nil {
				return
			}

			// And finish off by reading theirs.
			var remotePubKey [32]byte
			if _, err := io.ReadFull(c, remotePubKey[:]); err != nil {
				return
			}

			srwc := NewSecureReadWriteCloser(c, privateKey, &remotePubKey)

			for {
				io.Copy(srwc, srwc)
			}
		}(conn)
	}

}

func main() {
	port := flag.Int("l", 0, "Listen mode. Specify port")
	flag.Parse()

	// Server mode
	if *port != 0 {
		l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", *port))
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
	conn, err := Dial(os.Args[1])
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
