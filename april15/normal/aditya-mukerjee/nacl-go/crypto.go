package main

import (
	"crypto/rand"
	"fmt"
	"io"
	"math/big"

	"golang.org/x/crypto/nacl/box"
)

// Challenge assumes a maximum size for messages
const MaxMessageBytes = 32000

// NewSecureReader creates a new Reader. Reads from the returned Reader
// read, decrypt, and validate data from r. The private key is used to
// decrypt data, whereas the public key is used to validate the
// signature.
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	return &secureReader{priv: priv, pub: pub, r: r}
}

// NewSecureWriter creates a new Writer. Writes to the returned writer
// are encrypted with the provided public key, signed with the provided
// private key, and written to w.
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	return &secureWriter{priv: priv, pub: pub, w: w}
}

// A secureReader wraps an existing io.Reader which contains an encrypted
// and signed message. It decrypts it with the private key and
// authenticates it with the public key.
type secureReader struct {
	priv *[32]byte
	pub  *[32]byte
	r    io.Reader
}

func (sr *secureReader) Read(p []byte) (n int, err error) {
	dest := make([]byte, MaxMessageBytes)

	n, err = sr.r.Read(dest)
	if err != nil {
		return n, err
	}
	bts := dest[:n]

	nonce := [24]byte{}
	for i := range nonce {
		nonce[i] = bts[len(bts)-24+i]
	}

	message := bts[:len(bts)-24]

	out := []byte{}
	result, valid := box.Open(out, message, &nonce, sr.pub, sr.priv)
	if !valid {
		// TODO figure out how many bytes to return
		return len(bts), fmt.Errorf("unsigned message -- beware!")
	}

	copy(p, result)

	return len(result), nil
}

// A secureWriter takes data written to it and writes the encrypted and
// signed version to the underlying Writer (see NewSecureWriter)
type secureWriter struct {
	priv *[32]byte
	pub  *[32]byte
	w    io.Writer
}

func (sw *secureWriter) Write(p []byte) (n int, err error) {
	nonce, err := rand24Bytes()
	if err != nil {
		return 0, err
	}
	message := p

	var out []byte
	result := box.Seal(out, message, nonce, sw.pub, sw.priv)

	// The last 24 bytes of a message will be the random nonce
	result = append(result, nonce[:]...)
	n, err = sw.w.Write(result)
	if err == io.EOF {
		err = nil
	}
	return n, err
}

func rand24Bytes() (*[24]byte, error) {

	const Max = 256

	bts := &[24]byte{}
	for i := range bts {
		n, err := rand.Int(rand.Reader, big.NewInt(Max))
		if err != nil {
			return nil, err
		}
		if len(n.Bytes()) == 0 {
			continue
		}
		bts[i] = n.Bytes()[0]
	}
	return bts, nil
}
