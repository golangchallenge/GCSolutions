package main

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"

	"golang.org/x/crypto/nacl/box"
)

// ErrMsgTooLarge is the error emitted when Write is provided with a
// message that is too large for transmission.
var ErrMsgTooLarge = errors.New("message too large")

// SecureWriter is an io.Reader that writes encrypted messages.  It must
// be created by calling NewSecureWriter in order to provide decryption
// keys.
type SecureWriter struct {
	w   io.Writer // underlying writer
	key [32]byte  // cached shared key
	out []byte    // preallocated buffer for message encryption
}

// NewSecureWriter creates a new SecureWriter.
// w is the underlying transport.
// priv is our private key and pub is the other end's public key.
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	s := &SecureWriter{w: w, out: make([]byte, 0, 1024)}
	box.Precompute(&s.key, pub, priv)
	return s
}

// SecureWriter.Write writes the plaintext p to the underlying
// io.Writer. len(p) must be at most MaxMsgLen bytes.
func (s *SecureWriter) Write(p []byte) (int, error) {
	if len(p) > MaxMsgLen {
		return 0, ErrMsgTooLarge
	}

	// prepare header
	var header struct {
		DataLen int32
		Nonce   [nonceLen]byte
	}

	if _, err := rand.Read(header.Nonce[:]); err != nil {
		return 0, err
	}

	s.out = box.SealAfterPrecomputation(s.out[:0], p, &header.Nonce, &s.key)
	header.DataLen = int32(len(s.out))

	// write header
	if err := binary.Write(s.w, binary.BigEndian, &header); err != nil {
		return 0, err
	}

	// write message
	if _, err := writeFull(s.w, s.out); err != nil {
		// report 0 bytes of p as written. It doesn't make sense
		// to report n since it is not true that p[:n] was
		// written. This error is not recoverable as we have
		// already written something to s.w, and the other end
		// is likely trying to read more bytes than what we
		// wrote.
		return 0, err
	}

	return len(p), nil
}
