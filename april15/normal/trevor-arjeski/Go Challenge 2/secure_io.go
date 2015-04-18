package main

import (
	"errors"
	"golang.org/x/crypto/nacl/box"
	"io"
	"math/rand"
	"time"
)

const nonceSize = 24

var seeded = false
var myNonce [nonceSize]byte
var peerNonce [nonceSize]byte

// Seed rand once.
func seed() {
	rand.Seed(time.Now().UTC().UnixNano())
	seeded = true
}

// Get a nonce. Not the best nonce, but a nonce.
func getNonce() [nonceSize]byte {
	if !seeded {
		seed()
	}
	var nonce [nonceSize]byte
	for i := range nonce {
		nonce[i] = uint8(rand.Intn(255))
	}
	return nonce
}

// A reader/writer/closer that is a wrapper
// for a secure Conn.
type SecureReadWriteCloser struct {
	srw io.ReadWriteCloser
	r   io.Reader
	w   io.Writer
	c   io.Closer
}

// NewSecureReadWriteCloser instantiates a new SecureReadWriteCloser
func NewSecureReadWriteCloser(r io.Reader, w io.Writer, c io.Closer) io.ReadWriteCloser {
	return &SecureReadWriteCloser{r: r, w: w, c: c}
}

// Doesn't do anything
func (srwc *SecureReadWriteCloser) Close() error {
	return nil
}

// Calls the SecureReadWriterCloser's SecureReader.Read()
func (srwc *SecureReadWriteCloser) Read(p []byte) (n int, err error) {
	n, err = srwc.r.Read(p)
	return
}

// Calls the SecureReadWriterCloser's SecureReader.Write()
func (srwc *SecureReadWriteCloser) Write(p []byte) (n int, err error) {
	n, err = srwc.w.Write(p)
	return
}

// An implentation of io.Reader that has
// a private and public keypair.
type SecureReader struct {
	r         io.Reader
	priv, pub *[32]byte
}

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	return &SecureReader{r: r, priv: priv, pub: pub}
}

// Read() implentation for SecureReader.
//
// First this reads a nonce from the peer,
// then reads the encrypted message.
// Finally, the message is decrypted with the keypair and stored into
// the return buffer.
func (sr *SecureReader) Read(p []byte) (n int, err error) {
	// Read peer nonce
	n, err = sr.r.Read(peerNonce[:])
	buf := make([]byte, 32*1024)
	// Read encrypted message
	n, err = sr.r.Read(buf)
	// Open the box!
	opened, ok := box.Open(nil, buf[:n], &peerNonce, sr.pub, sr.priv)
	err = nil
	if !ok {
		err = errors.New("Could not decrypt message")
		return
	}
	//copy to output buffer
	n = copy(p, opened)
	return
}

// An implentation of io.Writer that has
// a private and public keypair.
type SecureWriter struct {
	w         io.Writer
	priv, pub *[32]byte
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	return &SecureWriter{w: w, priv: priv, pub: pub}
}

// Write() implentation for SecureWriter.
//
// First this generates a nonce to use for encryption,
// then reads the encrypted message.
// Finally, the message is encrypted with the keypair and stored into
// the return buffer.
func (sw *SecureWriter) Write(p []byte) (n int, err error) {
	// Get a new nonce to send
	myNonce = getNonce()
	var out []byte
	out = box.Seal(out, p, &myNonce, sw.pub, sw.priv)
	sw.w.Write(myNonce[:])
	sw.w.Write(out)
	return len(out), err
}

// An implentation of io.Closer.
type SecureCloser struct {
	c io.Closer
}

// NewSecureCloser instantiates a new SecureCloser
func NewSecureCloser(c io.Closer) io.Closer {
	return &SecureCloser{c: c}
}

// Doesn't do anything.
func (c *SecureCloser) Close() error {
	return nil
}
