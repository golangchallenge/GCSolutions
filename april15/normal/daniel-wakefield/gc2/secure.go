package main

import (
	"crypto/rand"
	"crypto/subtle"
	"fmt"
	"io"
	"io/ioutil"

	"golang.org/x/crypto/nacl/box"
)

var (
	nonce              = NewNonce()
	replaceReaderNonce = false
	replaceWriterNonce = false
	empty              = []byte{}
)

// SecureWriter holds the underlying communication channel and encryption keys
// for using it.
type SecureWriter struct {
	w     io.WriteCloser
	key   [32]byte
	nonce [24]byte
}

// SecureReader holds the underlying communication channel and encryption keys
// for using it.
type SecureReader struct {
	r     io.ReadCloser
	key   [32]byte
	nonce [24]byte
}

// SecureReadWriteCloser holds SecureReaders and SecureWriters together.
type SecureReadWriteCloser struct {
	SecureReader
	SecureWriter
}

// NewNonce sets up a nonce that is both random and sequential.
// This is useful against replay attacks.
// Using the least significant 6 bytes as the counter still leaves 2^36
// possible messages before a loop.
func NewNonce() [24]byte {
	b := [24]byte{}
	rand.Read(b[:18])
	return b
}

func incrNonce(n *[24]byte) {
	zero := 0

	// Will always do 6 comparisons and incrs so as to
	// not leak information on how large the nonce counter is.
	// Probably unnecessary but it cant hurt.
	for k, v := range [...]int{23, 22, 21, 20, 19, 18} {
		n[v]++
		// Check for byte wrap around from 255 to 0
		// If it occurs incr the next most significant byte
		if subtle.ConstantTimeByteEq(n[v], 0) != 1 {
			for i := 0; i < 5-k; i++ {
				subtle.ConstantTimeByteEq(0, 0)
				zero++
			}
			break
		}
	}

	zero = 0
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	wc, found := w.(io.WriteCloser)
	if !found {
		panic(fmt.Sprintf("Could not cast %v to io.WriteCloser", w))
	}

	s := SecureWriter{
		w:     wc,
		nonce: nonce,
	}

	// Ensure that successive writers will have different nonces while
	// a Reader, Writer pair will be created with the same nonce initially.
	if replaceWriterNonce {
		s.nonce = NewNonce()
	}
	replaceWriterNonce = true

	box.Precompute(&s.key, pub, priv)
	return s
}

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	rc, found := r.(io.ReadCloser)
	if !found {
		rc = ioutil.NopCloser(r)
	}

	s := SecureReader{
		r:     rc,
		nonce: nonce,
	}

	if replaceReaderNonce {
		s.nonce = NewNonce()
	}
	replaceReaderNonce = true

	box.Precompute(&s.key, pub, priv)
	return s
}

// NewSecureReadWriteCloser wraps a SecureReader and SecureWriter in the
// ReadWriteCloser interface while ensuring that they share a nonce.
func NewSecureReadWriteCloser(conn io.ReadWriter, priv, pub *[32]byte, nonce [24]byte) io.ReadWriteCloser {
	nsr := NewSecureReader(conn, priv, pub)
	nsw := NewSecureWriter(conn, priv, pub)

	sr := nsr.(SecureReader)
	sw := nsw.(SecureWriter)

	sr.nonce = nonce
	sw.nonce = nonce

	s := SecureReadWriteCloser{
		sr,
		sw,
	}

	return s
}

// Write encrypts b and writes it to the embedded Writer
func (s SecureWriter) Write(b []byte) (n int, err error) {
	msg := box.SealAfterPrecomputation(empty, b, &s.nonce, &s.key)

	n, err = s.w.Write(msg)
	if err != nil {
		return n, err
	}

	incrNonce(&s.nonce)
	return n, nil
}

// Read reads from the embedded Reader and decrypts the msg to b.
func (s SecureReader) Read(b []byte) (n int, err error) {
	msg := make([]byte, len(b)+box.Overhead)

	n, err = s.r.Read(msg)
	if err != nil {
		return n, err
	}

	out, ok := box.OpenAfterPrecomputation(empty, msg[:n], &s.nonce, &s.key)
	if !ok {
		panic("Could not decrypt")
	}

	copy(b, out)

	incrNonce(&s.nonce)
	return len(out), nil
}

// Close calls the Close methods on the embedded Reader and Writer.
func (s SecureReadWriteCloser) Close() (err error) {
	if err = s.SecureReader.Close(); err != nil {
		return err
	}
	if err = s.SecureWriter.Close(); err != nil {
		return err
	}
	return nil
}

// Close closes the Reader stream.
func (s SecureReader) Close() (err error) {
	if err = s.r.Close(); err != nil {
		return err
	}
	return nil
}

// Close closes the Writer stream.
func (s SecureWriter) Close() (err error) {
	if err = s.w.Close(); err != nil {
		return err
	}
	return nil
}
