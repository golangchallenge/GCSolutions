package main

import (
	"errors"
	"io"

	"golang.org/x/crypto/nacl/box"
)

// Errors returned by SecureReader
var (
	ErrInvalidMessage = errors.New("message could not be decrypted")
	ErrOpenBox        = errors.New("could not open box")
)

// A SecureReader is an io.Reader that can be read to retrieve decrypted
// data from an underlying encrypted stream.
//
// SecureReader expects to be able to read at least the minimum number
// of bytes to successfully decrypt data. If io.EOF is returned from
// the underlying stream before the minimum number of bytes is read,
// SecureReader returns io.ErrUnexpectedEOF.
//
// SecureReader uses nacl to authenticate and decrypt messages. It
// expects the first 24 bytes of the underlying stream to contain a
// nonce, and the remaining bytes to contain the ciphertext.
//
type SecureReader struct {
	r        io.Reader
	key      [ksize]byte
	nonce    [nsize]byte
	buf, msg []byte
	n        int
}

// NewSecureReader creates a new SecureReader to read from r.
func NewSecureReader(r io.Reader, priv, pub *[ksize]byte) io.Reader {
	// Allocate a single buffer for reading into, and a msg slice for
	// decrypting into. msg can also be used to store the remaining
	// stream on short reads.
	sr := &SecureReader{
		r:   r,
		key: [ksize]byte{},
		buf: make([]byte, nsize+maxBoxSize),
		msg: make([]byte, maxDecSize),
	}
	box.Precompute(&sr.key, pub, priv)
	return sr
}

// Read reads the next len(p) bytes from the SecureReader or until the
// SecureReader's buffer is drained.
//
// In the case that the decrypted message does not fit into p, the
// remainder is made available on the next call to Read.
func (r *SecureReader) Read(p []byte) (n int, err error) {
	// Previous short read?
	if r.n > 0 && r.n < len(r.msg) {
		n = copy(p, r.msg[r.n:])
		r.n += n
		return n, err
	}

	nr, err := io.ReadAtLeast(r.r, r.buf, nsize+minBoxSize)
	if nr == 0 {
		return 0, io.EOF
	} else if err == io.ErrUnexpectedEOF {
		return 0, ErrInvalidMessage
	} else if err != nil {
		return 0, err
	}

	// The first nsize bytes of the stream are the nonce; the box then
	// follows.
	copy(r.nonce[:], r.buf[:nsize])

	var ok bool
	r.msg, ok = box.OpenAfterPrecomputation(r.msg[:0], r.buf[nsize:nr], &r.nonce, &r.key)
	n = copy(p, r.msg)
	if !ok {
		return n, ErrOpenBox
	}

	r.n += n
	return n, err
}
