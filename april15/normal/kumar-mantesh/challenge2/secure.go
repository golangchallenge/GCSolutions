package main

import (
	"crypto/rand"
	"errors"
	"io"

	"golang.org/x/crypto/nacl/box"
)

type SecureKeys struct {
	priv    *[keySize]byte
	peerPub *[keySize]byte
	share   *[keySize]byte
}

type SecureReader struct {
	SecureKeys
	r io.Reader
}

type SecureWriter struct {
	SecureKeys
	w io.Writer
}

type secureReadWriteCloser struct {
	io.Reader
	io.Writer
	io.Closer
}

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, peerPub *[keySize]byte) io.Reader {
	var share [keySize]byte
	box.Precompute(&share, peerPub, priv)
	keys := SecureKeys{priv, peerPub, &share}
	return SecureReader{keys, r}
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, peerPub *[keySize]byte) io.Writer {
	var share [keySize]byte
	box.Precompute(&share, peerPub, priv)
	keys := SecureKeys{priv, peerPub, &share}
	return SecureWriter{keys, w}
}

// Extended Read method for reader to read secure data
func (sr SecureReader) Read(p []byte) (n int, err error) {
	//get the nonce
	r := make([]byte, nonceSize)
	if _, err = sr.r.Read(r); err != nil {
		return
	}
	var nonce [nonceSize]byte
	copy(nonce[:], r)

	//fetch encrypted message
	enc := make([]byte, maxMessageSize)
	encLen, err := sr.r.Read(enc)
	if err != nil {
		return 0, nil
	}
	enc = enc[:encLen]

	//compute the length of decrypted message
	n = encLen - box.Overhead

	//decrypt
	var s bool
	dec := make([]byte, n)
	if dec, s = box.OpenAfterPrecomputation(nil, enc, &nonce, sr.SecureKeys.share); s != true {
		return 0, errors.New("Error opening the box")
	}

	copy(p, dec)

	return
}

// Extended Write method for SecureWriter to write secure data
func (sw SecureWriter) Write(p []byte) (n int, err error) {
	//generate nonce
	r := make([]byte, nonceSize)
	if _, err = rand.Read(r); err != nil {
		return
	}
	var nonce [nonceSize]byte
	copy(nonce[:], r)

	//encrypt
	sp := box.SealAfterPrecomputation(nil, p, &nonce, sw.SecureKeys.share)

	m := append(nonce[:], sp...)

	return sw.w.Write(m)
}
