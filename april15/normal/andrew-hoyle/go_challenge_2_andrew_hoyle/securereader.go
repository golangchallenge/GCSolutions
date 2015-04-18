package main

import (
	"crypto/rand"
	"io"

	"golang.org/x/crypto/nacl/box"
)

type SecureReader struct {
	r    io.Reader
	priv *[32]byte
	pub  *[32]byte
}

func makeNonce() [24]byte {
	var n [24]byte
	io.ReadFull(rand.Reader, n[:])
	return n
}

func (sr *SecureReader) Read(p []byte) (int, error) {
	msgBuf := make([]byte, 1024)
	var nonBuf [24]byte

	sr.r.Read(nonBuf[:])
	n, err := sr.r.Read(msgBuf)
	if err != nil {
		return 0, err
	}

	out, _ := box.Open(nil, msgBuf[:n], &nonBuf, sr.pub, sr.priv)
	copy(p, out)

	return len(out), err
}

func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	return &SecureReader{
		r:    r,
		priv: priv,
		pub:  pub,
	}
}
