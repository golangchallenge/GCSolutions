package main

import (
	"crypto/rand"
	"errors"
	"io"

	"golang.org/x/crypto/nacl/box"
)

type naclWriter struct {
	sharedKey *[keyLen]byte
	out       io.Writer
}

// Newsecurewriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, peersPub *[32]byte) io.Writer {
	return newSecureWriter(w, priv, peersPub)
}

func newSecureWriter(w io.Writer, priv, peersPub *[keyLen]byte) *naclWriter {
	nw := &naclWriter{
		sharedKey: &[keyLen]byte{},
		out:       w,
	}

	box.Precompute(nw.sharedKey, peersPub, priv)

	return nw
}

var randReader = rand.Reader

func newNonce() ([nonceLen]byte, error) {
	n := [nonceLen]byte{}
	_, err := io.ReadFull(randReader, n[:])
	return n, err
}

func (w *naclWriter) Write(buf []byte) (n int, err error) {
	nonce, err := newNonce()
	if err != nil {
		return 0, err
	}

	if len(zbuf) >= 1<<16 {
		return 0, errors.New("oversized write")
	}

	out := box.SealAfterPrecomputation(nil, buf, &nonce, w.sharedKey)

	head := naclMsgHeader{
		Nonce:  nonce,
		Length: uint16(len(out)),
	}
	err = binWrite(w.out, head)
	if err != nil {
		return 0, err
	}

	n, err = w.out.Write(out)
	if err != nil {
		return n, err
	}

	return len(buf), nil
}
