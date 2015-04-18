package sio

import "io"

// A SecureWriter will encrypt data for its io.Writer using the NaCL library.
//
// To properly initialize a new SecureWriter, it must be obtained through
// the NewSecureWriter function.
type SecureWriter struct {
	priv, pub *[32]byte
	w         io.Writer
}

// NewSecureWriter will initialize and return a SecureWriter ready for use as
// an io.Writer. The array priv is the local private key and pub is the
// peer's public key.
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) (sw *SecureWriter) {
	sw = new(SecureWriter)
	sw.w = w
	sw.priv = priv
	sw.pub = pub
	return sw
}

// Write encrypts the contents of p and writes the data to the SecureWriter's
// io.Writer. The return value n is the length of the encrypted message
// placed in p.
func (sw SecureWriter) Write(p []byte) (n int, err error) {
	var out []byte

	if out, err = pack(sw.priv, sw.pub, p); err != nil {
		return
	}

	if n, err = sw.w.Write(out); err != nil {
		return
	}

	n = len(p)

	return
}
