package drum

import "bytes"

// A .splice file signature
var spliceSignature = []byte("SPLICE\x00\x00\x00\x00\x00\x00\x00")

// spliceHeader is a representation of splice file header.
type spliceHeader struct {
	Signature []byte `splice:"13"`
	DataSize  byte
}

// SignatureValid checks whether file signature is valid.
func (h *spliceHeader) SignatureValid() bool {
	return bytes.Compare(h.Signature, spliceSignature) == 0
}
