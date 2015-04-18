package main

import (
	"io"

	"golang.org/x/crypto/nacl/box"
)

type SecureWriter struct {
	w    io.Writer
	priv *[32]byte
	pub  *[32]byte
}

func (sw *SecureWriter) Write(p []byte) (int, error) {
	non := makeNonce()
	out := box.Seal(nil, p, &non, sw.pub, sw.priv)

	_, err := sw.w.Write(non[:])
	if err != nil {
		return 0, err
	}

	return sw.w.Write(out)
}

func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	return &SecureWriter{
		w:    w,
		priv: priv,
		pub:  pub,
	}
}
