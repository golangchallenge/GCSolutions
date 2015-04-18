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

// maxMessageSize is the largest size in bytes that
// a SecureReader or SecureWriter can processes.
const maxMessageSize = 1024 * 32

// A SecureReader wraps an io.Reader, a shared key and
// size and position data for the underlaying reader.
//
// Size and position data are useful for possible future
// "chunked" encrypted/decrypt functionality.
type SecureReader struct {
	r         io.Reader
	sharedKey *[32]byte
	size      uint32
	pos       uint32
}

// Read reads up to len(p) bytes into p. Encrypted data is read
// from the underlying io.Reader and decrypted with the shared key.
//
// The entire message must be decrypted at once, so this method
// returns an error if the decrypted message is larger than len(p).
// It will also return an error if the decrypted message is
// larger than maxMessageSize.
func (r *SecureReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	message := make([]byte, len(p)+24+box.Overhead)
	n, err := r.r.Read(message)
	if err != nil && err != io.EOF {
		return 0, err
	}
	if n == 0 {
		return 0, nil
	}
	message = message[:n]

	// nonce is first 24 bytes of message
	var nonce [24]byte
	copy(nonce[:], message[:24])

	// size of encrypted message is first 4 bytes of nonce
	if r.pos == uint32(0) {
		var size uint32
		buf := bytes.NewReader(nonce[0:4])
		bErr := binary.Read(buf, binary.LittleEndian, &size)
		if bErr != nil {
			return 0, bErr
		}
		r.size = size
	}

	// find length of decrypted message
	messageSize := r.size - uint32(24+box.Overhead)

	if messageSize > uint32(len(p)) {
		return 0, errors.New("buffer is too small")
	} else if messageSize > uint32(maxMessageSize) {
		return 0, errors.New("message is too large")
	}

	r.pos += uint32(n)

	// decrypt message with shared key
	decrypted, ok := box.OpenAfterPrecomputation(nil, message[24:], &nonce, r.sharedKey)
	if !ok {
		return 0, errors.New("unable to open box")
	}
	copy(p, decrypted)

	return len(decrypted), err
}

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	var sharedKey [32]byte
	box.Precompute(&sharedKey, pub, priv)
	return &SecureReader{r, &sharedKey, uint32(0), uint32(0)}
}

// A SecureWriter wraps a io.Writer and a shared key.
type SecureWriter struct {
	w         io.Writer
	sharedKey *[32]byte
}

// Write encrypts and writes len(p) bytes from p to the
// underlying data stream.
func (w *SecureWriter) Write(p []byte) (int, error) {
	// validate the incoming message
	if len(p) == 0 {
		return 0, errors.New("nothing to write")
	} else if len(p) > maxMessageSize {
		return 0, errors.New("message is too large")
	}

	// calculate the size of the resulting encrypted message
	buf := new(bytes.Buffer)
	bSize := len(p) + 24 + box.Overhead
	err := binary.Write(buf, binary.LittleEndian, uint32(bSize))
	if err != nil {
		return 0, err
	}
	size := buf.Bytes()

	random := make([]byte, 20)
	if _, err := rand.Read(random); err != nil {
		return 0, err
	}

	// first 4 bytes of nonce are the encrypted message size
	// last 20 bytes are random
	var nonce [24]byte
	copy(nonce[0:4], size[:])
	copy(nonce[4:24], random[:])

	// use shared key to encrypt message
	encrypted := box.SealAfterPrecomputation(nonce[:], p, &nonce, w.sharedKey)

	w.w.Write(encrypted)
	return len(p), nil
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	var sharedKey [32]byte
	box.Precompute(&sharedKey, pub, priv)
	return &SecureWriter{w, &sharedKey}
}

// A SecureReadWriteCloser wraps a reader, a writer,
// and a closer. It implements io.ReadWriteCloser.
type SecureReadWriteCloser struct {
	sr io.Reader
	sw io.Writer
	net.Conn
}

// NewSecureReadWriteCloser instantiates a new SecureReadWriteCloser
func NewSecureReadWriteCloser(conn net.Conn, priv, pub *[32]byte) io.ReadWriteCloser {
	sr := NewSecureReader(conn, priv, pub)
	sw := NewSecureWriter(conn, priv, pub)

	return &SecureReadWriteCloser{sr, sw, conn}
}

// Read delegates to the underlying SecureReader
func (srwc SecureReadWriteCloser) Read(p []byte) (int, error) {
	return srwc.sr.Read(p)
}

// Write delegates to the underlying SecureWriter
func (srwc SecureReadWriteCloser) Write(p []byte) (int, error) {
	return srwc.sw.Write(p)
}

// GenerateKey generates a new public/private key pair suitable
// for use with SecureReadWriteCloser.
func GenerateKey() (publicKey, privateKey *[32]byte, err error) {
	return box.GenerateKey(rand.Reader)
}
