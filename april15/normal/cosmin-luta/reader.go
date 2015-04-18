//
// Go Challenge 2
//
// Cosmin Luță <q4break@gmail.com>
//
// reader.go - SecureReader, io.Reader which performs decryption
//

package main

import (
	"errors"
	"io"

	"golang.org/x/crypto/nacl/box"
)

// ErrDecryptionFailed indicates that the message decryption failed.
var ErrDecryptionFailed = errors.New("decryption failed")

// A SecureReader implements the io.Reader interface and performs automatic decryption.
type SecureReader struct {
	readNonce            bool // flag indicating if the nonce has been read
	nonce                [NonceSize]byte
	sharedPrecomputedKey [KeySize]byte
	r                    io.Reader // we read encrypted data from this io.Reader
	buf                  []byte
	decryptedData        []byte // storage for decrypted data
}

func (s *SecureReader) Read(b []byte) (n int, err error) {

	if !s.readNonce {
		// If the nonce hasn't been read yet, do it now
		if n, err = s.r.Read(s.nonce[:]); err != nil || n != len(s.nonce) {
			return 0, ErrDecryptionFailed
		}
		s.readNonce = true
	}

	if len(b) == 0 {
		// Attempted to read into a 0 length slice
		return 0, nil
	}

	if len(s.decryptedData) == 0 {
		// No decrypted data available, read some from s.r and decrypt it
		n, err = s.r.Read(s.buf)
		if err != nil && err != io.EOF {
			// don't exit on EOF, it will be handled when the returned number of bytes
			// is 0.
			return 0, err
		}

		if n == 0 {
			return 0, io.EOF
		}

		s.buf = s.buf[:n]

		var ok bool
		// Decrypt and authenticate data
		s.decryptedData, ok = box.OpenAfterPrecomputation(nil,
			s.buf,
			&s.nonce,
			&s.sharedPrecomputedKey)
		if !ok {
			return 0, ErrDecryptionFailed
		}
	}

	// Copy as much as possible from s.decryptedData to b. If not all fits,
	// it will be returned on the next Read() call.
	n = copy(b, s.decryptedData)
	s.decryptedData = s.decryptedData[n:]
	return n, nil
}

// NewSecureReader returns an initialized SecureReader structure.
func NewSecureReader(r io.Reader, priv, pub *[KeySize]byte) io.Reader {
	var s = SecureReader{
		r:   r,
		buf: make([]byte, MaxMessageSize),
	}
	// Precompute a shared key for speeding up operations
	box.Precompute(&s.sharedPrecomputedKey, pub, priv)
	return &s
}
