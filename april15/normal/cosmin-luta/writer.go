//
// Go Challenge 2
//
// Cosmin Luță <q4break@gmail.com>
//
// writer.go - SecureWriter, io.Writer which performs encryption
//

package main

import (
	"crypto/rand"
	"io"

	"golang.org/x/crypto/nacl/box"
)

// SecureWriter implements the io.Writer interface and performs automatic encryption
// of all data written using the Write() method.
type SecureWriter struct {
	wroteNonce           bool // if set, the nonce has been written
	nonce                [NonceSize]byte
	sharedPrecomputedKey [KeySize]byte // precomputed key for speeding up encryption
	w                    io.Writer
}

func (s *SecureWriter) Write(p []byte) (n int, err error) {

	if !s.wroteNonce {
		// Write the nonce, if this is the first call to Write()
		if _, err = s.w.Write(s.nonce[:]); err != nil {
			return 0, err
		}
		s.wroteNonce = true
	}

	if len(p) > 0 {
		// Encrypt the message
		d := box.SealAfterPrecomputation(nil, p, &s.nonce, &s.sharedPrecomputedKey)
		if _, err = s.w.Write(d); err != nil {
			return 0, err
		}
	}
	return len(p), nil
}

// NewSecureWriter creates a new SecureWriter.
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	var s SecureWriter

	// Generate a random nonce
	if _, err := rand.Read(s.nonce[:]); err != nil {
		return nil
	}
	s.w = w
	box.Precompute(&s.sharedPrecomputedKey, pub, priv)
	return &s
}
