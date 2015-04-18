package main

import (
	"encoding/binary"
	"errors"
	"io"

	"golang.org/x/crypto/nacl/box"
)

// ErrDecryptionError is the error emitted when decryption fails.
var ErrDecryptionError = errors.New("failed to decrypt input")

// SecureReader is an io.Reader that reads encrypted messages.  It must
// be created by calling NewSecureReader in order to provide decryption
// keys.
type SecureReader struct {
	r   io.Reader // underlying reader
	key [32]byte  // cached shared key
	in  []byte    // preallocated buffer for message transmission
}

// NewSecureReader creates a new SecureReader.
// r is the underlying transport.
// priv is our private key and pub is the other end's public key.
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	s := &SecureReader{r: r, in: make([]byte, 1024)}
	box.Precompute(&s.key, pub, priv)
	return s
}

// SecureReader.Read reads a message from the underlying io.Reader and
// places the plaintext in p, which must have enough capacity to hold
// it in its entirety.
//
// SecureReader diverges from the way io.Reader works in that it returns
// the number of bytes put into p, not the number of bytes read (which
// is always larger than the returned value).
//
// If the message fails to decrypt, ErrDecryptionError is returned.
func (s *SecureReader) Read(p []byte) (int, error) {
	var header struct {
		DataLen int32
		Nonce   [nonceLen]byte
	}

	if err := binary.Read(s.r, binary.BigEndian, &header); err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return 0, err
	}

	if header.DataLen > MaxMsgLen+box.Overhead {
		return 0, ErrMsgTooLarge
	}

	if int(header.DataLen) > cap(p)+box.Overhead {
		// p does not have enough space to hold the decrypted
		// message. Since we have already read the header, we
		// are pretty much doomed.  This error is not
		// recoverable.
		return 0, io.ErrShortBuffer
	}

	if cap(s.in) < int(header.DataLen) {
		s.in = make([]byte, header.DataLen)
	}
	s.in = s.in[:header.DataLen]
	if _, err := io.ReadFull(s.r, s.in); err != nil {
		return 0, err
	}

	p, ok := box.OpenAfterPrecomputation(p[:0], s.in, &header.Nonce, &s.key)
	if !ok {
		return 0, ErrDecryptionError
	}

	return len(p), nil
}
