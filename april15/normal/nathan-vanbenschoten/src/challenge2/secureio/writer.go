package secureio

import (
	"challenge2/secureio/nonce"
	"io"

	"golang.org/x/crypto/nacl/box"
)

// NewSecureWriter instantiates a new secureWriter that wraps an io.Writer
// and provides a client that encrypts all messages with the NaCL cryptography
// system.
//
// The writer uses a nonce that is appended to the front of all messages to
// prevent replay attacks. Coupled with encrypted messages using public-key
// cryptography, this assures secure communication.
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	return &secureWriter{w, priv, pub, nonce.NewNonce()}
}

// secureWriter wraps an io.Writer and contains both a peers public key and its
// own private key to use with public-key cryptography. The client also hold
// onto a nonce that is incremented and sent with each message to prevent
// replay attacks.
type secureWriter struct {
	w         io.Writer
	priv, pub *[32]byte
	n         nonce.Nonce
}

// Write sends the writer's current nonce, along with an encrypted form of
// the provided message to the secureWriter's contained io.Writer.
//
// After the write, the secureWriter's nonce will be incremented.
func (sw *secureWriter) Write(p []byte) (n int, err error) {
	defer sw.n.Increment()
	bytes := box.Seal(sw.n.Slice(), p, sw.n.Array(), sw.pub, sw.priv)
	return sw.w.Write(bytes)
}
