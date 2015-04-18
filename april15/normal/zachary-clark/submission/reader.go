package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"golang.org/x/crypto/nacl/box"
	"io"
)

// SecureReader is an implemenation of io.Reader that uses NaCl.
type SecureReader struct {
	reader io.Reader
	priv   *[32]byte
	pub    *[32]byte
}

type errReader struct {
	r   io.Reader
	err error
}

func (er *errReader) read(buf []byte) {
	if er.err != nil {
		return
	}
	_, er.err = er.r.Read(buf)
}

// Reads an encrypted message from the reader expecting the following structure:
// The first two bytes are the length(x) of the encrypted message.
// The next 24 bytes are the nonce.
// The next x bytes is the encrypted message.
func (s *SecureReader) Read(p []byte) (int, error) {
	er := &errReader{r: s.reader}

	plen := make([]byte, 2)
	er.read(plen)

	payloadLen, err := binary.ReadVarint(bytes.NewBuffer(plen))
	if err != nil {
		return 0, errors.New("Invalid length bytes")
	}

	nonce := make([]byte, 24)
	er.read(nonce)

	payload := make([]byte, payloadLen)
	er.read(payload)

	if er.err != nil {
		return 0, er.err
	}

	nonceArray := &[24]byte{}
	copy(nonceArray[:], nonce)

	message, success := box.Open(nil, payload, nonceArray, s.pub, s.priv)
	if !success {
		return 0, errors.New("Decrypt failure")
	}

	copy(p, message)
	return len(message), nil
}

// NewSecureReader instantiates a new SecureReader.
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	return &SecureReader{r, priv, pub}
}
