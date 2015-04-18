package main

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"golang.org/x/crypto/nacl/box"
	"io"
)

const (
	maxLength   = 32000
	nonceLength = 24
	sizeLength  = 2
)

// SecureReadWriter implements encrypting ReadWriter
type SecureReadWriter struct {
	shared *[32]byte

	reader io.Reader
	writer io.Writer
}

// NewSecureReadWriter is a SecureReadWriter constructor
func NewSecureReadWriter(priv, pub *[32]byte, r io.Reader, w io.Writer) *SecureReadWriter {
	var shared [32]byte
	sr := &SecureReadWriter{&shared, r, w}
	box.Precompute(sr.shared, pub, priv)
	return sr
}

// Read encrypted data
func (srw *SecureReadWriter) Read(b []byte) (int, error) {
	if srw.reader == nil {
		return 0, fmt.Errorf("reader not set")
	}
	scratch := make([]byte, len(b)+box.Overhead+nonceLength)

	n, err := srw.reader.Read(scratch[:sizeLength])
	if err != nil {
		return 0, err
	}
	if n != sizeLength {
		return 0, fmt.Errorf("failed to read size")
	}
	pSize, _ := binary.Uvarint(scratch[:sizeLength])
	size := int(pSize)

	if len(b) < size-nonceLength-box.Overhead {
		return 0, fmt.Errorf("buffer is too small")
	}

	var read int
	for read < size {
		n, err = srw.reader.Read(scratch[read:])
		if err != nil {
			return 0, err
		}
		read += n
	}

	var nonce [nonceLength]byte
	copy(nonce[:], scratch[:nonceLength])

	res := make([]byte, 0, read-nonceLength-box.Overhead)
	out, ok := box.OpenAfterPrecomputation(res, scratch[nonceLength:read], &nonce, srw.shared)
	if !ok {
		return 0, fmt.Errorf("failed decrypting")
	}
	copy(b, out)

	return len(out), nil
}

// Write encrypted data
func (srw *SecureReadWriter) Write(b []byte) (int, error) {
	if srw.writer == nil {
		return 0, fmt.Errorf("writer not set")
	}

	if len(b) > maxLength {
		return 0, fmt.Errorf("data is too big")
	}

	buf := make([]byte, sizeLength+nonceLength, len(b)+box.Overhead+sizeLength+nonceLength)

	size := uint64(len(b)) + nonceLength + box.Overhead
	binary.PutUvarint(buf[:sizeLength], size)

	var nonce [nonceLength]byte
	_, err := rand.Read(nonce[:])
	if err != nil {
		return 0, err
	}
	copy(buf[sizeLength:], nonce[:])

	buf = box.SealAfterPrecomputation(buf, b, &nonce, srw.shared)

	_, err = srw.writer.Write(buf)
	if err != nil {
		return 0, err
	}
	return len(b), nil
}
