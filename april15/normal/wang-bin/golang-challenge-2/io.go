package main

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"golang.org/x/crypto/nacl/box"
	"io"
)

// MaxMsgLength is the maxium length of bytes could be write once.
const MaxMsgLength = 32768

// Errors returned by SecureReader and SecureWriter.
var (
	ErrMsgTooLong = errors.New("message too long")
	ErrDecrypt    = errors.New("failed to decrypt message")
)

// SecureWriter implements NaCl encryption for an io.Writer object.
type SecureWriter struct {
	w         io.Writer
	priv, pub *[32]byte
	err       error
}

func (sw *SecureWriter) nonce() *[24]byte {
	nonce := new([24]byte)
	_, sw.err = rand.Read(nonce[:])
	return nonce
}

func (sw *SecureWriter) write(bo binary.ByteOrder, v interface{}) {
	if sw.err != nil {
		return
	}
	sw.err = binary.Write(sw.w, bo, v)
}

// Write encrypts p using NaCl's box package and writes into the buffer.
// It returns the length of p if no error happens, otherwise n will be 0.
func (sw *SecureWriter) Write(p []byte) (int, error) {
	if len(p) > MaxMsgLength {
		return 0, ErrMsgTooLong
	}
	nonce := sw.nonce()
	enc := box.Seal(nil, p, nonce, sw.pub, sw.priv)
	sw.write(binary.LittleEndian, int32(len(enc)))
	sw.write(binary.LittleEndian, enc)
	sw.write(binary.LittleEndian, nonce)
	if sw.err != nil {
		return 0, sw.err
	}
	return len(p), nil
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	return &SecureWriter{w: w, priv: priv, pub: pub}
}

// SecureReader implements NaCl decryption for an io.Reader object.
type SecureReader struct {
	r         io.Reader
	priv, pub *[32]byte
	err       error
}

func (sr *SecureReader) read(bo binary.ByteOrder, v interface{}) {
	if sr.err != nil {
		return
	}
	sr.err = binary.Read(sr.r, bo, v)
}

// Read reads data, decrypted use NaCl's box package and writes into p.
// It returns the number of bytes read into p.
func (sr *SecureReader) Read(p []byte) (n int, err error) {
	var size int32
	sr.read(binary.LittleEndian, &size)
	encryptedMsg := make([]byte, int(size))
	sr.read(binary.LittleEndian, encryptedMsg)
	nonce := new([24]byte)
	sr.read(binary.LittleEndian, nonce)
	if sr.err != nil {
		return 0, sr.err
	}
	decryptedMsg, ok := box.Open(nil, encryptedMsg, nonce, sr.pub, sr.priv)
	if !ok {
		return 0, ErrDecrypt
	}
	n = copy(p, decryptedMsg[:])
	return n, nil
}

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	return &SecureReader{r: r, priv: priv, pub: pub}
}

func generateKey() (pub *[32]byte, priv *[32]byte, err error) {
	pub, priv, err = box.GenerateKey(rand.Reader)
	return
}
