package sio

import (
	"errors"
	"io"
)

// A SecureReader will decrypt data from its io.Reader using the NaCL library
//
// To properly initialize a new SecureReader, it must be obtained through
// the NewSecureReader function
type SecureReader struct {
	priv, pub *[32]byte
	r         io.Reader
}

// NewSecureReader will initialize and return a SecureReader ready for use as
// an io.Reader. The array priv is the local private key and pub is the
// peer's public key.
func NewSecureReader(r io.Reader, priv, pub *[32]byte) (sr *SecureReader) {
	sr = new(SecureReader)
	sr.r = r
	sr.priv = priv
	sr.pub = pub
	return
}

// Read reads data from the SecureReader's io.Reader, decrypts it and copy()s
// it to p. The return value n is the length of the decrypted message placed
// in p.
func (sr SecureReader) Read(p []byte) (n int, err error) {
	var out []byte

	if len(p) == 0 {
		return 0, errors.New("secureReader.Read() received 0 length buffer")
	}

	data := make([]byte, len(p))
	if n, err = sr.r.Read(data); err != nil {
		return
	}

	if out, err = unpack(sr.priv, sr.pub, n, data); err != nil {
		return
	}

	if n = copy(p, out); n < len(out) {
		return n, errors.New("secureReader.Read() had " + string(len(out)-n) + " more bytes than buffer")
	}

	return
}
