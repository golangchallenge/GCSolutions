package main

import (
	"crypto/rand"
	"errors"
	"io"
	"net"

	"golang.org/x/crypto/nacl/box"
)

// Number of bytes in a private or public key
const KeyLength = 32

// Number of bytes in a nonce
const nonceLength = 24

// Max length of a message including the nonce and overhead
const maxMsgLen = 32*1024 + nonceLength + box.Overhead

// GenerateNonce generates a unique nonce.
//
// @return [[]byte] random bytes
func GenerateNonce() ([nonceLength]byte, error) {
	rb := make([]byte, nonceLength)
	_, err := rand.Read(rb)
	if err != nil {
		return [nonceLength]byte{}, err
	}

	var result [nonceLength]byte
	for i, x := range rb {
		result[i] = x
	}

	return result, nil
}

// GenerateKeyPair generates a private/public key pair for use in sekurity
//
// @return [[]byte] public key
// @return [[]byte] private key
// @return [Error]
func GenerateKeyPair() (publicKey, privateKey *[KeyLength]byte, err error) {
	return box.GenerateKey(rand.Reader)
}

// SecureReader conforms to the io.Reader interface
type SecureReader struct {
	reader    io.Reader
	sharedKey *[KeyLength]byte
}

// Read encrypted data from the secure reader.
//
// @param p [[]byte] buffer to store decrypted message in
// @return [Integer] number of decrypted bytes read
// @return [Error] if error, or nil
func (secureReader SecureReader) Read(p []byte) (int, error) {
	// Get the encrypted bytes from the internal reader
	data := make([]byte, maxMsgLen)
	n, err := secureReader.reader.Read(data)
	if n == 0 || (err != nil && err != io.EOF) {
		return 0, err
	}

	//  The first set of bytes is the nonce followed by the message.
	var nonce [nonceLength]byte
	for i := 0; i < nonceLength; i++ {
		nonce[i] = data[i]
	}

	decryptedData, success := box.OpenAfterPrecomputation(nil, data[nonceLength:n], &nonce, secureReader.sharedKey)
	if success != true {
		return 0, errors.New("Unable to decrypt message.")
	}

	// Copy results into p
	msgLen := n - nonceLength - box.Overhead
	for i := 0; i < msgLen; i++ {
		p[i] = decryptedData[i]
	}

	return msgLen, nil
}

// SecureWriter conforms to the io.Writer interface
type SecureWriter struct {
	writer    io.Writer
	sharedKey *[KeyLength]byte
}

// Write encrypted data to the secure writer.
//
// @param p [[]byte] buffer of bytes to encrypt and write
// @return [Integer] number of decrypted bytes written
// @return [Error] if error, or nil
func (secureWriter SecureWriter) Write(p []byte) (int, error) {
	// Nonce must be unique per message for security
	nonce, err := GenerateNonce()
	if err != nil {
		return 0, err
	}

	data := box.SealAfterPrecomputation(nonce[:], p, &nonce, secureWriter.sharedKey)
	n, err := secureWriter.writer.Write(data)
	if err != nil {
		return 0, err
	}

	// Unencrypted bytes written does not include the overhead
	msgLen := n - nonceLength - box.Overhead
	return msgLen, nil
}

// SecureReaderWriter conforms to the io.ReadWriteCloser interface
type SecureReaderWriter struct {
	conn         net.Conn // Retain conn to close later
	secureReader io.Reader
	secureWriter io.Writer
}

// Read reads and decrypts bytes bytes from connection into buffer p
//
// @param p [[]byte] buffer to put decrypted data in
// @return n [int] number of bytes read
// @return err [Error]
func (rw SecureReaderWriter) Read(p []byte) (n int, err error) {
	return rw.secureReader.Read(p)
}

// Write encrypts and writes bytes in buffer p to connection
//
// @param p [[]byte] buffer to write to connection
// @return n [int] number of bytes written
// @return err [Error]
func (rw SecureReaderWriter) Write(p []byte) (n int, err error) {
	return rw.secureWriter.Write(p)
}

// Close closes the secure reader-writer connection
//
// @return [Error]
func (rw SecureReaderWriter) Close() error {
	return rw.conn.Close()
}

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, pub *[KeyLength]byte) io.Reader {
	sharedKey := new([KeyLength]byte)
	box.Precompute(sharedKey, pub, priv)
	return SecureReader{
		r,
		sharedKey,
	}
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[KeyLength]byte) io.Writer {
	sharedKey := new([KeyLength]byte)
	box.Precompute(sharedKey, pub, priv)
	return SecureWriter{
		w,
		sharedKey,
	}
}

// NewSecureReaderWriter instantiates a new SecureReaderWriter
func NewSecureReaderWriter(conn net.Conn, priv, pub *[KeyLength]byte) io.ReadWriteCloser {
	secureReader := NewSecureReader(conn, priv, pub)
	secureWriter := NewSecureWriter(conn, priv, pub)

	return SecureReaderWriter{
		conn,
		secureReader,
		secureWriter,
	}
}
