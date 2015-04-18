// Package boxcrypto was created for the golang-challenge number 2
// (http://golang-challenge.com/go-challenge2/) by:
// Adrian Fletcher
// github.com/AdrianFletcher
// adrian@fletchtechnology.com.au

// Package boxcrypto is a basic wrapper of the nacl/box package. Crypto will produce
// a unique nonce for each message and prepend the nonce to each call to Write.
//
// This package has no knowledge of the underlying networking/messenging platform
// and does not perform any protocols or handshakes. The client should
// implement these themselves.
package boxcrypto

import (
	"crypto/rand"
	"fmt"
	"golang.org/x/crypto/nacl/box"
	"io"
)

// KeyPair contains the public and private keys neccesary for encryption/decryption
type KeyPair struct {
	private *[32]byte
	public  *[32]byte
}

// CryptoReader wraps a Public/Private keypair and implements the io.Reader interface
type CryptoReader struct {
	reader io.Reader
	kp     *KeyPair
}

// CryptoWriter wraps a Public/Private keypair and implements the io.Reader interface
type CryptoWriter struct {
	writer io.Writer
	kp     *KeyPair
}

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	return &CryptoReader{
		reader: r,
		kp:     &KeyPair{priv, pub},
	}
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	return &CryptoWriter{
		writer: w,
		kp:     &KeyPair{priv, pub},
	}
}

// Read will attempt to decrypt any message from the in-built Reader.
// the first 24 bytes of the incoming message are expected to be
// the correct nonce to use to decrypt.
func (cr *CryptoReader) Read(p []byte) (n int, err error) {
	// The maximum expected size of message is 32 kilobytes
	buf := make([]byte, 32*1024)

	n, err = cr.reader.Read(buf)
	if err != nil {
		return 0, fmt.Errorf("Could not read from stream. %s", err)
	}

	// Nonce must be a 24 byte array for the nacl Open method
	if n < 24 {
		return 0, fmt.Errorf("Nonce was %d bytes. Must be 24 bytes.", n)
	}
	var nonce [24]byte
	copy(nonce[:], buf[:24])

	result, ok := box.Open(nil, buf[24:n], &nonce, cr.kp.public, cr.kp.private)
	if ok != true {
		return 0, fmt.Errorf("Could not decrypt message.")
	}

	copy(p, result)
	return len(result), nil
}

// Write will encrypt the provided message, prepending a new random nonce to
// the message before being sent. The public and private 'keys' on your secureWriter
// should be generate before Writing to the stream.
func (cw *CryptoWriter) Write(p []byte) (n int, err error) {
	nonce, err := getNonce()
	if err != nil {
		return 0, fmt.Errorf("Could not generate nonce: %s.", err)
	}

	out := box.Seal(nonce[:], p, &nonce, cw.kp.public, cw.kp.private)

	n, err = cw.writer.Write(out)
	if err != nil {
		return 0, fmt.Errorf("Could not write to stream: %s.", err)
	}

	return n, nil
}

// GenerateKey will create a matching cryptographically secure public and
// private key of 32 bytes in length.
func GenerateKeyPair() (publicKey, privateKey *[32]byte, err error) {
	return box.GenerateKey(rand.Reader)
}

// getNonce will generate a new, cryptographically secure 24bit nonce. If the
// generated nonce is not 24 bits exactly, it will return an error.
func getNonce() ([24]byte, error) {
	var nonce [24]byte
	n, err := rand.Read(nonce[:24])
	if err != nil {
		return nonce, fmt.Errorf("Could not read from rand.Read(): %s.", err)
	}
	if n != 24 {
		return nonce, fmt.Errorf("Generated nonce is not 24 bytes")
	}
	return nonce, nil
}
