// Package main implements readers and writers that utilizes NaCL to securely
// encrypt/decrypt messages.
package main

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"
	"net"

	"golang.org/x/crypto/nacl/box"
)

// NewSecureWriter returns a new SecureWriter that writes messages
// to w encrypted with the provided private/public key pair.
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	shared := &[32]byte{}
	box.Precompute(shared, pub, priv)
	return secureWriter{w, shared}
}

type secureWriter struct {
	pipe   io.Writer
	shared *[32]byte
}

// Write encrypts b using the private/public key pair
// and writes the encrypted message to the underlying data stream in the
// following format:
//
//	description				offset	type
//	----------------------------------------
//	body size (nonce+box)	0		uint16
//	nonce					2		[24]byte
//	box						26		[]byte
//
// It returns the number of post-encrypted bytes written and any error encountered
func (s secureWriter) Write(b []byte) (int, error) {
	// generate nonce
	nonce := [24]byte{}
	_, err := rand.Read(nonce[:])
	if err != nil {
		return 0, err
	}

	// encrypt input and append ciphertext to nonce
	carton := box.SealAfterPrecomputation(nonce[:], b, &nonce, s.shared)

	// write header
	err = binary.Write(s.pipe, binary.LittleEndian, uint16(len(carton)))
	if err != nil {
		return 0, err
	}
	// write body
	n, err := s.pipe.Write(carton)
	return 2 + n, err
}

// NewSecureReader returns a new SecureReader that reads messages
// from r decrypted with the provided private/public key pair.
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	shared := &[32]byte{}
	box.Precompute(shared, pub, priv)
	return secureReader{r, shared}
}

type secureReader struct {
	pipe   io.Reader
	shared *[32]byte
}

// Read reads a message from the underlying data stream and decrypts it using
// the private/public key pair, which is then copied up to len(b) bytes into b
// Read returns the number of decrypted bytes read (0 <= n <= len(p)) and any
// error encountered.
func (s secureReader) Read(b []byte) (int, error) {
	// parse header
	var size uint16
	err := binary.Read(s.pipe, binary.LittleEndian, &size)
	if err != nil {
		return 0, err
	}

	// parse body
	var buf bytes.Buffer
	_, err = buf.ReadFrom(io.LimitReader(s.pipe, int64(size)))
	nonce := [24]byte{}
	copy(nonce[:], buf.Next(24))
	cipherText := buf.Bytes()

	// decrypt body and copy to buffer
	msg, success := box.OpenAfterPrecomputation(nil, cipherText, &nonce, s.shared)
	if !success {
		return 0, errors.New("unable to decrypt cipher text. box.Open() returned false")
	}
	return copy(b, msg), nil
}

// newSecureConnection returns a new SecureConnection that uses a SecureWriter
// to perform write operations and a SecureReader to perform read operations
func newSecureConnection(conn net.Conn, priv, pub *[32]byte) io.ReadWriteCloser {
	return secureConnection{conn, NewSecureReader(conn, priv, pub), NewSecureWriter(conn, priv, pub)}
}

type secureConnection struct {
	conn   net.Conn
	reader io.Reader
	writer io.Writer
}

func (s secureConnection) Write(b []byte) (int, error) {
	return s.writer.Write(b)
}

func (s secureConnection) Read(b []byte) (int, error) {
	return s.reader.Read(b)
}

func (s secureConnection) Close() error {
	return s.conn.Close()
}

// newErrReadWriter returns a new ErrReadWriter that implements the sticky
// error pattern. When an error is encountered, ErrReadWriter caches
// the error and performs no-op for subsequent invocations.
func newErrReadWriter(rw io.ReadWriter) errReadWriter {
	return errReadWriter{rw, rw, 0, nil}
}

type errReadWriter struct {
	r   io.Reader
	w   io.Writer
	n   int
	err error
}

func (e *errReadWriter) write(b []byte) int {
	if e.err != nil {
		return 0
	}
	e.n, e.err = e.w.Write(b)
	return e.n
}

func (e *errReadWriter) read(b []byte) int {
	if e.err != nil {
		return 0
	}
	e.n, e.err = e.r.Read(b)
	return e.n
}
