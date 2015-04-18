package main

import (
	"errors"
	"io"

	"golang.org/x/crypto/nacl/box"
)

var (
	ErrBadKey = errors.New("could not decrypt")
)

const (
	nonceLen  = 24
	keyLen    = 32
	lengthLen = 2
)

type naclMsgHeader struct {
	Nonce  [nonceLen]byte
	Length uint16
}

type naclReader struct {
	sharedKey *[keyLen]byte
	in        io.Reader
	leftovers []byte
}

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, peersPub *[32]byte) io.Reader {
	return newSecureReader(r, priv, peersPub)
}

func newSecureReader(r io.Reader, priv, peersPub *[32]byte) *naclReader {
	n := &naclReader{
		sharedKey: &[keyLen]byte{},
		in:        r,
	}

	box.Precompute(n.sharedKey, peersPub, priv)
	return n
}

func (r *naclReader) Read(buf []byte) (n int, err error) {

	if len(r.leftovers) != 0 {
		n = copy(buf, r.leftovers)
		r.leftovers = r.leftovers[n:]
		return n, nil
	}

	head := naclMsgHeader{}
	err = binRead(r.in, &head)
	if err != nil {
		return 0, err
	}

	cipherText := make([]byte, head.Length)
	_, err = io.ReadFull(r.in, cipherText)
	if err != nil {
		return 0, err
	}

	msg, ok := box.OpenAfterPrecomputation(nil, cipherText, &head.Nonce, r.sharedKey)
	if !ok {
		return 0, ErrBadKey
	}

	n = copy(buf, msg)
	if len(msg) != n {
		r.leftovers = msg[n:]
	}

	return n, nil
}
