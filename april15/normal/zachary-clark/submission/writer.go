package main

import (
	"crypto/rand"
	"encoding/binary"
	"golang.org/x/crypto/nacl/box"
	"io"
)

// SecureWriter is an implementation of io.Writer that uses NaCl.
type SecureWriter struct {
	writer io.Writer
	priv   *[32]byte
	pub    *[32]byte
}

type errWriter struct {
	w   io.Writer
	err error
}

func (ew *errWriter) write(buf []byte) {
	if ew.err != nil {
		return
	}
	_, ew.err = ew.w.Write(buf)
}

func generateNonce() (*[24]byte, error) {
	rbytes := make([]byte, 24)
	_, err := rand.Read(rbytes)
	nonce := &[24]byte{}
	copy(nonce[:], rbytes)
	return nonce, err
}

// Encrypts and writes to the writer outputing the following structure:
// The first two bytes are the length(x) of the encrypted message.
// The next 24 bytes are a random nonce.
// The next x bytes is the encrypted message.
func (s *SecureWriter) Write(p []byte) (int, error) {
	nonce, err := generateNonce()
	if err != nil {
		return 0, err
	}
	payload := box.Seal(nil, p, nonce, s.pub, s.priv)

	plen := make([]byte, 2)
	binary.PutVarint(plen, int64(len(payload)))

	ew := &errWriter{w: s.writer}
	ew.write(plen)
	ew.write(nonce[:])
	ew.write(payload)

	if ew.err != nil {
		// If an error occurred, bytes may have been written
		// but they do not correspond to bytes of `p` written,
		// which is what the io.Writer interface cares about.
		// SecureWriter is pretty much all or nothing.
		return 0, ew.err
	}

	return len(p), ew.err
}

// NewSecureWriter instantiates a new SecureWriter.
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	return &SecureWriter{w, priv, pub}
}
