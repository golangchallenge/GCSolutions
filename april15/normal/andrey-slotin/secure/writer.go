package secure

import (
	"crypto/rand"
	"io"

	"golang.org/x/crypto/nacl/box"
)

// NonceLenth is a size of byte slice required by NaCl.
const (
	NonceLength = 24
)

// Writer wraps io.Writer and encrypts its content using NaCl.
type Writer struct {
	writer    io.Writer
	sharedKey [32]byte
	n         uint64
}

// NewWriter instantiates a new Writer.
func NewWriter(w io.Writer, priv, pub *[32]byte) *Writer {
	sw := &Writer{
		writer: w,
	}

	box.Precompute(&sw.sharedKey, pub, priv)

	return sw
}

func (sw *Writer) Write(p []byte) (n int, err error) {
	out := make([]byte, 0, len(p)+box.Overhead+NonceLength)
	nonce := sw.generateNonce()
	out = box.SealAfterPrecomputation(out, p, &nonce, &sw.sharedKey)
	out = append(out, nonce[:]...)

	return sw.writer.Write(out)
}

func (sw *Writer) generateNonce() (nonce [NonceLength]byte) {
	rand.Reader.Read(nonce[:])

	return nonce
}
