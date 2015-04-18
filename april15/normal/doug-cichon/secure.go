package main

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"
	"net"

	"golang.org/x/crypto/nacl/box"
)

// DecryptionError is returned if there is an error while decrypting
var ErrWhileDecrypting = errors.New("decryption error")

// SecureWriter will encrypt plaintext and write ciphertext to writer.
type SecureWriter struct {
	writer     io.Writer
	publicKey  *[32]byte
	privateKey *[32]byte
}

// NewSecureWriter instantiates a new SecureWriter.
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	return SecureWriter{writer: w, privateKey: priv, publicKey: pub}
}

// Write will encrypt p, then write the ciphertext.
// Data is in the format: nonce (24 bytes), cipherlength (2 bytes), ciphertext (cipherlength+box.Overhead bytes)
func (w SecureWriter) Write(p []byte) (int, error) {
	var nonce [24]byte
	rand.Read(nonce[:])

	// Write the plaintext nonce to the stream
	err := binary.Write(w.writer, binary.BigEndian, &nonce)
	if err != nil {
		return 0, err
	}

	// Write the length of the plaintext message
	err = binary.Write(w.writer, binary.BigEndian, uint16(len(p)))
	if err != nil {
		return 0, err
	}

	c := box.Seal(nil, p, &nonce, w.publicKey, w.privateKey)

	err = binary.Write(w.writer, binary.BigEndian, c)
	if err != nil {
		return 0, err
	}

	// Return the number of plaintext bytes written, the user is not concerned with the overhead
	return len(c) - box.Overhead, nil
}

// SecureReader will read encrypted data, turning it into plaintext.
type SecureReader struct {
	reader     io.Reader
	publicKey  *[32]byte
	privateKey *[32]byte
}

// NewSecureReader instantiates a new SecureReader.
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	return SecureReader{reader: r, privateKey: priv, publicKey: pub}
}

// Read will read ciphertext from reader, and put plaintext into slice p.
// See Write documentation for details on the data format.
func (r SecureReader) Read(p []byte) (int, error) {
	// Read the plaintext nonce from the stream
	var nonce [24]byte
	err := binary.Read(r.reader, binary.BigEndian, &nonce)
	if err != nil {
		return 0, err
	}

	// Read message length from the stream.
	// Since the longest expected message is 32KB, only two bytes are necessary.
	var length uint16
	err = binary.Read(r.reader, binary.BigEndian, &length)
	if err != nil {
		return 0, err
	}

	// Read the ciphertext from the stream.
	// The ciphertext will be the message length, plus the amount of overhead.
	ct := make([]byte, length+box.Overhead)
	err = binary.Read(r.reader, binary.BigEndian, &ct)
	if err != nil {
		return 0, err
	}

	// Decrypt the message.
	opened, ok := box.Open(nil, ct, &nonce, r.publicKey, r.privateKey)
	if !ok {
		return 0, ErrWhileDecrypting
	}

	copy(p, opened)

	return len(opened), nil
}

// SecureConnection is a connection which has encrypted communication.
type SecureConnection struct {
	conn net.Conn // The connection to wrap, it must be stored to call Close.
	SecureReader
	SecureWriter
}

// NewSecureConnection instantiates a new SecureConnection.
func NewSecureConnection(c net.Conn, priv, pub *[32]byte) SecureConnection {
	return SecureConnection{
		SecureReader: NewSecureReader(c, priv, pub).(SecureReader),
		SecureWriter: NewSecureWriter(c, priv, pub).(SecureWriter),
		conn:         c,
	}
}

// Close will close the underlying connection.
func (s SecureConnection) Close() error {
	return s.conn.Close()
}
