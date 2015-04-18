//This file contains the implementations of SecureWriter and SecureReader.
//They are used to transparently encrypt and decrypt data that goes through them,
//see http://golang-challenge.com/go-challenge2/
package main

import (
	"crypto/rand"
	"fmt"
	"golang.org/x/crypto/nacl/box"
	"io"
)

//These constants define the sizes of different parts of the messages
const (
	NonceLength  = 24
	HeaderLength = box.Overhead + NonceLength
	MaxMsgLength = 32768
	Bufsize      = HeaderLength + MaxMsgLength
)

//The SecureReader wraps an ordinary io.Reader and
//decrypts the bytes read from the underlying io.Reader
//using a public and private keypair defined on initialization
//It also utilizes a buffer to hold the cyphertext and facilitates
//messages up to 32KB (defined by Bufsize constant)
type SecureReader struct {
	r          io.Reader
	priv       *[32]byte
	pub        *[32]byte
	receivebuf [Bufsize]byte
	nonce      [24]byte
}

//The SecureWriter wraps an ordinary io.Writer and
//encrypts all bytes before sending them to the underlying writer
//using a public and private keypair defined on initialization.
//It facilitates the sending of messages up to 32KB (defined by Bufsize constant)
type SecureWriter struct {
	w       io.Writer
	priv    *[32]byte
	pub     *[32]byte
	sendbuf [Bufsize]byte
	nonce   [24]byte
}

//A wrapper type to securely send data through a socket
type SecureSocket struct {
	io.Reader
	io.Writer
	io.Closer
}

type MaxBytesError struct{}

func (MaxBytesError) Error() string {
	return fmt.Sprintf("Cannot operate on more than %d bytes at a time", MaxMsgLength)
}

//Reads an encrypted message from a SecureWriter and returns the
//unencrypted bytes into buf.
func (sr *SecureReader) Read(buf []byte) (n int, err error) {
	if len(buf) > MaxMsgLength {
		return 0, MaxBytesError{}
	}

	//Try to read the header and enough bytes to fill the output buffer
	n, err = sr.r.Read(sr.receivebuf[:len(buf)+HeaderLength])
	if err != nil {
		return n, err
	}
	copy(sr.nonce[:], sr.receivebuf[:24])
	cyphertext := sr.receivebuf[NonceLength:n]

	box.Open(buf[0:0], cyphertext, &sr.nonce, sr.pub, sr.priv)

	return n - HeaderLength, nil
}

//Allows sending encrypted bytes to a SecureReader as messages.
//Generates a random "nonce" on each invocation and prepends
//that to the message. Encrypts the bytes in buf, utilizing this nonce
//before sending them together with the "nonce" to the underlying writer.
func (sw *SecureWriter) Write(buf []byte) (n int, err error) {
	if len(buf) > MaxMsgLength {
		return 0, MaxBytesError{}
	}

	_, err = rand.Read(sw.nonce[:])
	copy(sw.sendbuf[:24], sw.nonce[:])
	box.Seal(sw.sendbuf[:24], buf, &sw.nonce, sw.pub, sw.priv)

	n, err = sw.w.Write(sw.sendbuf[:HeaderLength+len(buf)])
	if err != nil {
		return n, err
	}
	return n - HeaderLength, nil
}

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	sr := SecureReader{
		r:    r,
		priv: priv,
		pub:  pub,
	}

	return &sr
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	sw := SecureWriter{
		w:    w,
		priv: priv,
		pub:  pub,
	}

	return &sw
}
