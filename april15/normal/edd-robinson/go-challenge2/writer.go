package main

import (
	"crypto/rand"
	"io"

	"golang.org/x/crypto/nacl/box"
)

// A SecureWriter is an io.Writer that encrypts a plaintext underlying
// stream.
//
// SecureWriter uses nacl to authenticate and encrypt plaintext
// messages.
//
type SecureWriter struct {
	w     io.Writer
	key   [ksize]byte
	nonce [nsize]byte
	buf   []byte
}

// NewSecureWriter creates a new SecureWriter that writes to w.
func NewSecureWriter(w io.Writer, priv, pub *[ksize]byte) io.Writer {
	sw := &SecureWriter{
		w:   w,
		key: [ksize]byte{},
		buf: make([]byte, maxBoxSize),
	}
	box.Precompute(&sw.key, pub, priv)
	return sw
}

// Write writes an encrypted form of p to the underlying io.Writer.
//
// The first 24 bytes written are a random nonce. The remainder of the
// stream is the encrypted ciphertext of p.
//
// Because Write uses a new nonce on each call, Write will report having
// written 0 bytes of p if there is an error, or len(p) bytes if the
// encrypted form of p was successfully written to the underlying
// io.Writer.
func (w *SecureWriter) Write(p []byte) (n int, err error) {
	// rand.Reader will always provide len(w.nonce) bytes if err is nil.
	if _, err = rand.Reader.Read(w.nonce[:]); err != nil {
		return n, err
	}
	w.buf = box.SealAfterPrecomputation(w.buf[:0], p, &w.nonce, &w.key)

	// It is tricky to know whether it's more effective to append the
	// nonce and buffer together and write once, or as done here,
	// have two writes but save an allocation. In a real system you'd
	// want to figure out what's most efficient.
	if _, err = w.w.Write(w.nonce[:]); err != nil {
		return 0, err
	}
	if _, err = w.w.Write(w.buf); err != nil {
		return 0, err
	}
	return len(p), nil
}
