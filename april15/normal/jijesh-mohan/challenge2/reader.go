package main

import (
	"encoding/binary"
	"errors"
	"io"

	"golang.org/x/crypto/nacl/box"
)

var (
	ErrInvalidSize      = errors.New("decrypt: invalid size for data")
	ErrDecryptionFailed = errors.New("decrypt: decryption failed")
)

// SecureReader implements NACL secure reader which
// will decrypt the entrypted message when reading.
type SecureReader struct {
	r         io.Reader
	sharedKey *[32]byte
	err       error
	buf       []byte
}

// readNonce reads nonce from the reader.
func (sr *SecureReader) readNonce() *[24]byte {
	var nonce [24]byte
	_, sr.err = io.ReadFull(sr.r, nonce[:])
	return &nonce
}

// readLength reads encrypted message length from the reader.
func (sr *SecureReader) readLength() (length uint64) {
	if sr.err != nil {
		return
	}
	sr.err = binary.Read(sr.r, binary.LittleEndian, &length)
	return
}

// readBytes reads bytes with provided size from the reader.
func (sr *SecureReader) readBytes(size uint64) []byte {
	if sr.err != nil {
		return nil
	}
	data := make([]byte, size)
	_, sr.err = io.ReadFull(sr.r, data)
	return data
}

// Read reads encrypted bytes from the reader and decrypt.
func (sr *SecureReader) Read(p []byte) (n int, err error) {
	sr.err = nil

	if len(p) == 0 {
		return 0, ErrInvalidSize
	}
	// copy the buffer data if it is not empty
	if len(sr.buf) > 0 {
		n = copy(p, sr.buf)
		sr.buf = sr.buf[n:]
		if n == len(p) {
			return n, nil
		}
		p = p[n:]
	}

	nonce := sr.readNonce()
	length := sr.readLength()
	boxed := sr.readBytes(length)

	if sr.err != nil {
		return 0, sr.err
	}

	message, ok := box.OpenAfterPrecomputation(nil, boxed, nonce, sr.sharedKey)

	if !ok {
		return 0, ErrDecryptionFailed
	}

	n += copy(p, message)

	// store the remaining msg in the buffer.
	if len(message) > len(p) {
		sr.buf = message[n:]
	}

	return n, nil
}

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, privateKey, peersPublicKey *[32]byte) io.Reader {
	var sharedKey [32]byte
	box.Precompute(&sharedKey, peersPublicKey, privateKey)
	return &SecureReader{r: r, sharedKey: &sharedKey}
}
