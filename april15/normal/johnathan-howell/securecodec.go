package main

import (
	"crypto/rand"
	"errors"
	"golang.org/x/crypto/nacl/box"
	"io"
)
const (
	// We consider that our messages will always be smaller than 32KB.
	rcvbuf_sz = 32000
)

// SecureCodec implements a secure io.ReadWriteCloser
type SecureCodec struct {
	r      SecureReader
	w      SecureWriter
	closer io.Closer
}

// SecureReader implements a secure io.Reader
type SecureReader struct {
	r io.Reader
	// key is the shared key between the keypair passed to NewSecureReader computed by box.Precompute
	key *[32]byte
	rcvbuf []byte
}

// SecureWriter implements a secure io.Writer
type SecureWriter struct {
	w io.Writer
	// key is the shared key between the keypair passed to NewSecureWriter computed by box.Precompute
	key *[32]byte
}

func (s SecureCodec) Read(p []byte) (int, error) {
	n, err := s.r.Read(p)
	return n, err
}
func (s SecureCodec) Write(p []byte) (int, error) {
	n, err := s.w.Write(p)
	return n, err
}

// Close closes the underlying net.Conn, return err
func (s SecureCodec) Close() error {
	err := s.closer.Close()
	return err
}
func (s SecureReader) Read(p []byte) (int, error) {
	// Read ciphertext from SecureReader's underlying io.Reader
	n, err := s.r.Read(s.rcvbuf)
	if err != nil {
		return n, err
	}
	// If we read less than 24 bytes (the size of a nonce), the ciphertext cannot possibly be valid.
	if n < 24 {
		return n, errors.New("INVALID CIPHERTEXT")
	}
	// First 24 bytes of cipher are the Nonce
	var nonce [24]byte
	copy(nonce[:], s.rcvbuf[:24])

	// call box.Open on the ciphertext
	plaintext, success := box.OpenAfterPrecomputation(nil, s.rcvbuf[24:n], &nonce, s.key)
	if success != true {
		return 0, errors.New("DECRYPTION ERROR")
	}
	// Copy the newly decrypted plaintext into p
	copy(p, plaintext)
	return len(plaintext), err
}
func (s SecureWriter) Write(p []byte) (int, error) {
	// Randomly generate a Nonce and store it in the first 24 bytes of the ciphertext
	var nonce [24]byte
	_, err := rand.Read(nonce[:])
	if err != nil {
		return 0, err
	}
	// Call box.Seal to encrypt p using our keys
	ciphertext := box.SealAfterPrecomputation(nil, p, &nonce, s.key)
	// Append the Nonce to the beginning of the ciphertext
	ciphertext = append(nonce[:], ciphertext...)
	// Write the ciphertext into SecureWriter's underlying io.Writer
	n, err := s.w.Write(ciphertext)
	return n, err
}

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, pub *[32]byte) SecureReader {
	// Precompute the shared key between priv, pub
	var sharedkey [32]byte
	box.Precompute(&sharedkey, pub, priv)
	return SecureReader{r: r, key: &sharedkey, rcvbuf: make([]byte, rcvbuf_sz) }
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) SecureWriter {
	var sharedkey [32]byte
	box.Precompute(&sharedkey, pub, priv)
	return SecureWriter{w: w, key: &sharedkey}
}

// NewSecureCodec instantiates a new SecureCodec
func NewSecureCodec(r SecureReader, w SecureWriter, closer io.Closer) io.ReadWriteCloser {
	return SecureCodec{w: w, r: r, closer: closer}
}
