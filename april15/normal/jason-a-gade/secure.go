// secure.go contains types and routines for implementing
// NaCl-encrypted streams and network connections.

package main

import (
	"crypto/rand"
	"io"
	"net"

	"golang.org/x/crypto/nacl/box"
	"golang.org/x/crypto/nacl/secretbox"
)

const (
	keySize   = 32
	nonceSize = 24
)

// SecureReader implements NaCl decryption over an io.Reader stream.
type SecureReader struct {
	io.Reader
	key *[keySize]byte
}

// NewSecureReader returns a new SecureReader from r with a shared key formed
// from the private, public key pair.
func NewSecureReader(r io.Reader, priv, pub *[keySize]byte) SecureReader {
	s := SecureReader{Reader: r}
	s.key = new([keySize]byte)
	box.Precompute(s.key, pub, priv)
	return s
}

func (s SecureReader) Read(p []byte) (int, error) {
	n, err := s.Reader.Read(p)
	if err != nil {
		return 0, err
	}
	return copy(p, decrypt(p[:n], s.key)), nil
}

// decrypt returns a byte slice decrypted with key. If the input cannot
// be decrypted, it returns its input unchanged.
func decrypt(in []byte, key *[keySize]byte) []byte {
	if len(in) < nonceSize {
		return in
	}
	var nonce [nonceSize]byte
	copy(nonce[:], in)
	out, ok := secretbox.Open(nil, in[nonceSize:], &nonce, key)
	if !ok {
		return in
	}
	return out
}

// SecureWriter implements NaCl encryption over an io.Writer stream.
type SecureWriter struct {
	io.Writer
	key *[keySize]byte
}

// NewSecureWriter returns a new SecureWriter to w with a shared key formed
// from the private, public key pair.
func NewSecureWriter(w io.Writer, priv, pub *[keySize]byte) SecureWriter {
	s := SecureWriter{Writer: w}
	s.key = new([keySize]byte)
	box.Precompute(s.key, pub, priv)
	return s
}

func (s SecureWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	return s.Writer.Write(encrypt(p, s.key))
}

// encrypt returns a byte slice encrypted with key.
func encrypt(in []byte, key *[keySize]byte) []byte {
	nonce := newNonce()
	head := make([]byte, nonceSize)
	copy(head, nonce[:])
	return secretbox.Seal(head, in, nonce, key)
}

func newNonce() *[nonceSize]byte {
	var nonce [nonceSize]byte
	_, err := io.ReadFull(rand.Reader, nonce[:])
	if err != nil {
		panic(err)
	}
	return &nonce
}

// SecureConn implements NaCl-encryption and decryption over a net.Conn.
type SecureConn struct {
	net.Conn
	SecureReader
	SecureWriter
}

// NewSecureConn returns a new SecureConn using an existing net.Conn, and a new
// SecureReader and SecureWriter.
func NewSecureConn(c net.Conn, priv, peer *[keySize]byte) SecureConn {
	s := SecureConn{Conn: c}
	s.SecureReader = NewSecureReader(c, priv, peer)
	s.SecureWriter = NewSecureWriter(c, priv, peer)
	return s
}

func (s SecureConn) Read(p []byte) (int, error) {
	return s.SecureReader.Read(p)
}

func (s SecureConn) Write(p []byte) (int, error) {
	return s.SecureWriter.Write(p)
}
