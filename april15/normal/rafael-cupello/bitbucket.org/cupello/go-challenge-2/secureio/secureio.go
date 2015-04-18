package secureio

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"

	"golang.org/x/crypto/nacl/box"
)

// secureReader is the struct that wraps a reader and a sharedKey.
type secureReader struct {
	r         io.Reader
	sharedKey *[keySize]byte
}

// secureWriter is the struct that wraps a writer and a sharedKey.
type secureWriter struct {
	w         io.Writer
	sharedKey *[keySize]byte
}

// secureReadWriter is the struct that wraps a reader and a writer.
type secureReadWriter struct {
	r io.Reader
	w io.Writer
}

// secureReadWriteCloser is the struct that wraps a reader, a writer and a closer.
type secureReadWriteCloser struct {
	r io.Reader
	w io.Writer
	c io.Closer
}

const (
	// size of secret key
	keySize = 32

	// size of nonce
	nonceSize = 24
)

// generateNonce generates a random nonce.
func generateNonce() (*[nonceSize]byte, error) {
	nonce := new([nonceSize]byte)
	if _, err := rand.Read(nonce[:]); err != nil {
		return nil, err
	}
	return nonce, nil
}

// Read decrypts the message from the data stream and
// reads the result into buff.
func (sr secureReader) Read(buff []byte) (int, error) {

	// Read the 2 firsts bytes that carry the length of the message
	message := make([]byte, 2)
	if _, err := sr.r.Read(message); err != nil {
		return 0, err
	}
	length := binary.BigEndian.Uint16(message[:2])

	// Validate if message has a valid length
	if length < (nonceSize + box.Overhead) {
		return 0, errors.New("Failed to decrypt")
	}

	// Read the encrypted message
	message = make([]byte, length)
	ln, err := sr.r.Read(message)
	if err != nil {
		return 0, err
	}

	// Get the nonce from the message
	nonce := new([nonceSize]byte)
	copy(nonce[:], message[:nonceSize])

	// Decrypt the message.
	out, opened := box.OpenAfterPrecomputation(nil, message[nonceSize:ln], nonce, sr.sharedKey)
	if !opened {
		return 0, errors.New("Failed to decrypt")
	}

	// Copy the message decrypted to buff
	ln = copy(buff, out[:])

	// Return the length of the message copied to buff
	return ln, nil
}

// Write encrypts the message and
// writes the result to the underlying data stream.
func (s secureWriter) Write(message []byte) (int, error) {

	// Generate a random nonce
	nonce, err := generateNonce()
	if err != nil {
		return 0, errors.New("Failed to encrypt")
	}

	// Prepend the nonce to out
	out := make([]byte, nonceSize)
	copy(out[:nonceSize], nonce[:])

	// Encrypt the message and prepend to out
	out = box.SealAfterPrecomputation(out, message[:], nonce, s.sharedKey)

	// Prepend the length of the message in Uint16 to out's beginning
	size := make([]byte, 2)
	binary.BigEndian.PutUint16(size[:], uint16(len(out)))
	out = append(size, out...)

	// Write the message decrypted to out
	ln, err := s.w.Write(out)
	if err != nil {
		return 0, err
	}

	// Return the length of the message wrote to buff
	return ln, nil
}

// Write calls the Write method on the secureWriter object wrapped by secureReadWriter
func (srw secureReadWriter) Write(p []byte) (int, error) {
	return srw.w.Write(p)
}

// Read calls the Read method on the secureReader object wrapped by secureReadWriter
func (srw secureReadWriter) Read(p []byte) (int, error) {
	return srw.r.Read(p)
}

// Close calls the Close method on the Closer object wrapped by secureReadWriteCloser
func (srwc secureReadWriteCloser) Close() error {
	return srwc.c.Close()
}

// Write calls the Write method on the secureWriter object wrapped by secureReadWriteCloser
func (srwc secureReadWriteCloser) Write(p []byte) (int, error) {
	return srwc.w.Write(p)
}

// Read calls the Read method on the secureReader object wrapped by secureReadWriteCloser
func (srwc secureReadWriteCloser) Read(p []byte) (int, error) {
	return srwc.r.Read(p)
}

// GenerateKeyPair generates a new key pair using nacl/box.
func GenerateKeyPair() (pub, priv *[keySize]byte, err error) {
	pub, priv, err = box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	return pub, priv, nil
}

// NewsecureReader generates a sharedKey and instantiates a new secureReader
func NewSecureReader(r io.Reader, priv, peerPub *[32]byte) io.Reader {
	sharedKey := new([keySize]byte)
	box.Precompute(sharedKey, peerPub, priv)
	return secureReader{r: r, sharedKey: sharedKey}
}

// NewsecureWriter generates a sharedKey and instantiates a new secureWriter
func NewSecureWriter(w io.Writer, priv, peerPub *[32]byte) io.Writer {
	sharedKey := new([keySize]byte)
	box.Precompute(sharedKey, peerPub, priv)
	return secureWriter{w: w, sharedKey: sharedKey}
}

// NewsecureReadWriter instantiates a new secureReadWriter
func NewSecureReadWriter(rw io.ReadWriter, priv, peerPub *[32]byte) io.ReadWriter {
	return secureReadWriter{r: NewSecureReader(rw, priv, peerPub), w: NewSecureWriter(rw, priv, peerPub)}
}

// NewsecureReadWriteCloser instantiates a new secureReadWriteCloser
func NewSecureReadWriteCloser(rwc io.ReadWriteCloser, priv, peerPub *[32]byte) io.ReadWriteCloser {
	return secureReadWriteCloser{r: NewSecureReader(rwc, priv, peerPub), w: NewSecureWriter(rwc, priv, peerPub), c: rwc}
}
