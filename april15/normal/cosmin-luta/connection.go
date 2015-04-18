//
// Go Challenge 2
//
// Cosmin Luță <q4break@gmail.com>
//
// connection.go - Add encryption to a net.Conn
//

package main

import (
	"errors"
	"io"
	"net"
)

var (
	// ErrKeyExchangeFailed indicates that the key exchange failed.
	ErrKeyExchangeFailed = errors.New("key exchange failed")
	// ErrInitializationFailed indicates that there was a problem initializing a component.
	ErrInitializationFailed = errors.New("initialization error")
)

// EncryptedConnection is an io.ReadWriteCloser which transparently performs
// encryption and decryption.
type EncryptedConnection struct {
	reader io.Reader
	writer io.Writer
	closer io.Closer
}

// Read receives data while performing transparent decryption.
func (e *EncryptedConnection) Read(b []byte) (int, error) {
	return e.reader.Read(b)
}

// Write sends data while performing transparent encryption.
func (e *EncryptedConnection) Write(b []byte) (int, error) {
	return e.writer.Write(b)
}

// Close terminates the encrypted connection.
func (e *EncryptedConnection) Close() error {
	return e.closer.Close()
}

// NewEncryptedConnection performs the key exchange over the specified connection
// and returns an EncryptedConnection.
func NewEncryptedConnection(c net.Conn) (*EncryptedConnection, error) {

	// Generate a new key pair
	keyPair, err := NewKeyPair()
	if err != nil {
		return nil, err
	}

	// Send the public key to the peer
	if _, err = c.Write(keyPair.pub[:]); err != nil {
		return nil, ErrKeyExchangeFailed
	}

	var buf [KeySize]byte
	var n int

	// Attempt to read the peer's public key
	n, err = c.Read(buf[:])
	if err != nil && err != io.EOF {
		return nil, err
	}

	if n != KeySize {
		// Check if the key is exactly the expected size
		return nil, ErrKeyExchangeFailed
	}

	// Create a secure reader on top of the connection, for decrypting incoming data
	r := NewSecureReader(c, keyPair.priv, &buf)
	if r == nil {
		return nil, ErrInitializationFailed
	}

	// Create a secure reader on top of the connection, for encrypting outoging data
	w := NewSecureWriter(c, keyPair.priv, &buf)
	if w == nil {
		return nil, ErrInitializationFailed
	}

	return &EncryptedConnection{
		reader: r,
		writer: w,
		closer: c}, nil
}
