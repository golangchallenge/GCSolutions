package main

import (
	"crypto/rand"
	"encoding/binary"
	"io"

	"golang.org/x/crypto/nacl/box"
)

// SecureWriter implements NACL secure writer
// which encrypt the data before writing.
type SecureWriter struct {
	w         io.Writer
	sharedKey *[32]byte
	err       error
}

// Write writes encrypted message to the stream.
func (sw *SecureWriter) Write(p []byte) (int, error) {
	sw.err = nil
	nonce := sw.generateNonce()

	boxed := box.SealAfterPrecomputation(nil, p, nonce, sw.sharedKey)

	sw.writeBytes(nonce[:])
	sw.err = binary.Write(sw.w, binary.LittleEndian, uint64(len(boxed)))

	sw.writeBytes(boxed)
	if sw.err != nil {
		return 0, sw.err
	}
	return len(p), nil
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, privateKey, peersPublicKey *[32]byte) io.Writer {
	var sharedKey [32]byte
	box.Precompute(&sharedKey, peersPublicKey, privateKey)
	return &SecureWriter{w: w, sharedKey: &sharedKey}
}

// writeBytes writes bytes to underlying io.writer
func (sw *SecureWriter) writeBytes(p []byte) {
	if sw.err != nil {
		return
	}
	_, sw.err = sw.w.Write(p)
}

// generateNonce generate a random nonce.
func (sw *SecureWriter) generateNonce() *[24]byte {
	var nonce [24]byte
	_, sw.err = io.ReadFull(rand.Reader, nonce[:])
	return &nonce
}
