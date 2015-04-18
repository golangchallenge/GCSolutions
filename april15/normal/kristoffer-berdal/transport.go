package main

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"

	"golang.org/x/crypto/nacl/box"
)

// SecureReadWriteCloser implements io.ReadWriteCloser
type SecureReadWriteCloser struct {
	secureR io.Reader
	secureW io.Writer
	conn    io.Closer
}

// Read reads from the underlying SecureReader
func (rwc SecureReadWriteCloser) Read(p []byte) (n int, err error) {
	return rwc.secureR.Read(p)
}

// Write writes to the underlying SecureWriter
func (rwc SecureReadWriteCloser) Write(p []byte) (n int, err error) {
	return rwc.secureW.Write(p)
}

// Close closes the underlying net.Conn
func (rwc SecureReadWriteCloser) Close() error {
	return rwc.conn.Close()
}

// A SecureReader wraps a io.Reader to provide NaCl-encrypted communications
type SecureReader struct {
	reader         io.Reader
	privateKey     *[32]byte
	peersPublicKey *[32]byte
}

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	return &SecureReader{r, priv, pub}
}

// A SecureWriter wraps a io.Writer to provide NaCl-encrypted communications
type SecureWriter struct {
	writer         io.Writer
	privateKey     *[32]byte
	peersPublicKey *[32]byte
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	return &SecureWriter{w, priv, pub}
}

// getNonce generates a random [24]byte nonce using crypto/rand
func getNonce() (*[24]byte, error) {
	var n [24]byte
	_, err := rand.Read(n[:])
	if err != nil {
		return nil, err
	}
	return &n, nil
}

// Write writes the NaCl-encrypted data of p to the underlying datastream.
// See http://godoc.org/io#Writer for how Write(p []byte) works in general.
func (w *SecureWriter) Write(p []byte) (n int, err error) {
	// First we need to encrypt this, luckily we have a shaker in our toolbox, let's add some salt!
	var message []byte
	nonce, err := getNonce()
	if err != nil {
		return 0, err
	}
	message = box.Seal(nil, p, nonce, w.peersPublicKey, w.privateKey)
	ms := int64(binary.Size(message))
	// Write message length
	err = binary.Write(w.writer, binary.LittleEndian, ms)
	if err != nil {
		return 0, err
	}
	// Write nonce
	err = binary.Write(w.writer, binary.LittleEndian, nonce)
	if err != nil {
		return 0, err
	}
	// Write message
	err = binary.Write(w.writer, binary.LittleEndian, message)
	if err != nil {
		return 0, err
	}
	return int(ms), err
}

// Read reads and decrypts NaCl-encrypted data into p from the underlying datastream.
// See http://godoc.org/io#Reader for how Read(p []byte) works in general.
func (r *SecureReader) Read(p []byte) (n int, err error) {
	var ds int64
	err = binary.Read(r.reader, binary.LittleEndian, &ds)
	if err != nil {
		return 0, err
	}
	if ds == 0 {
		return 0, nil
	}
	var nonce [24]byte
	// Read the nonce sent over the wire.
	err = binary.Read(r.reader, binary.LittleEndian, &nonce)
	if err != nil {
		return 0, err
	}

	// Make a slice ds bytes big
	data := make([]byte, ds)
	// Read ds bytes from the connection into data
	err = binary.Read(r.reader, binary.LittleEndian, &data)
	if err != nil {
		return 0, err
	}
	// Now we need to decrypt the message.
	var message []byte
	message, ok := box.Open(nil, data, &nonce, r.peersPublicKey, r.privateKey)

	if !ok {
		return 0, errors.New("could not decrypt message")
	}
	// The message decrypted ok, now we read the data into a bytes buffer, this way we can get the length in a proper way.
	b := bytes.NewReader(message)
	i, err := b.Read(p)
	// We could check err here, but the error handling here would be the exact same as below, unless we want to print a specific message.
	return i, err
}
