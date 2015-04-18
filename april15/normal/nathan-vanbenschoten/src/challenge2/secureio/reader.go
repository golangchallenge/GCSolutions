package secureio

import (
	"challenge2/secureio/nonce"
	"fmt"
	"io"

	"golang.org/x/crypto/nacl/box"
)

// NewSecureReader instantiates a new secureReader that wraps an io.Reader
// and provides a client that is able to unencrypt all messages with the NaCL
// cryptography system.
//
// The reader holds a reference to the last valid nonce seen and compares this
// to the nonce provided on each message to prevent replay attacks. Coupled
// with encrypted messages using public-key cryptography, this assures secure
// communication.
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	return &secureReader{r, priv, pub, nil}
}

// secureReader wraps an io.Reader and contains both a peers public key and its
// own private key to use with public-key cryptography. The client also hold
// onto it's most recently validated nonce to prevent replay attacks.
type secureReader struct {
	r         io.Reader
	priv, pub *[32]byte
	last      nonce.Nonce
}

// Read reads in a message from the secureReader's contained io.Reader, first
// validating the provided nonce and then unencypting the rest of the message.
//
// After the read, the secureReader's most recently validated nonce will be
// updated.
func (sr *secureReader) Read(p []byte) (n int, err error) {
	in := make([]byte, 24+len(p)+box.Overhead)
	num, err := sr.r.Read(in)
	if err != nil {
		return 0, err
	}

	readNonce, err := nonce.FromBytes(in[:24])
	if err != nil {
		return 0, err
	}
	if sr.last != nil && !readNonce.After(sr.last) {
		return 0, fmt.Errorf("nonce read is not greater than previous nonce, aborting to avoid replay attack")
	}
	sr.last = readNonce

	bytes, ok := box.Open(nil, in[24:num], readNonce.Array(), sr.pub, sr.priv)
	if !ok {
		return 0, fmt.Errorf("there was an error")
	}
	return copy(p, bytes), nil
}
