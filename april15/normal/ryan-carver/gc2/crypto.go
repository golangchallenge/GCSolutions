package main

import (
	"crypto/rand"
	"fmt"
	"io"

	"golang.org/x/crypto/nacl/box"
)

const keySize = 32

// KeyPair holds a public/private key pair, and facilitiates performing
// a Diffie-Hellman key exchange.
type KeyPair struct {
	pub  *[keySize]byte
	priv *[keySize]byte
}

// NewKeyPair returns a new KeyPair initialized with public and private keys.
func NewKeyPair() *KeyPair {
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil
	}
	return &KeyPair{pub, priv}
}

// Exchange performs a key exchange over the ReadWriter. It first writes the
// public key to the Writer, then gets the peer's public key by reading from
// the Reader. A new KeyPair is returned containing the peer's public key and
// this private key.
func (kp KeyPair) Exchange(rw io.ReadWriter) (*KeyPair, error) {
	if err := kp.send(rw); err != nil {
		return nil, err
	}
	return kp.recv(rw)
}

func (kp *KeyPair) send(w io.Writer) error {
	debugf("Sending public key %v\n", kp.pub)
	if _, err := w.Write(kp.pub[:]); err != nil {
		return err
	}
	debugf("Sent public key %v\n", kp.pub)
	return nil
}

func (kp KeyPair) recv(r io.Reader) (*KeyPair, error) {
	newPair := &KeyPair{pub: &[keySize]byte{}, priv: kp.priv}
	debugf("Receiving...\n")
	if _, err := r.Read(newPair.pub[:]); err != nil {
		return nil, err
	}
	debugf("Received peer's public key: %v\n", newPair.pub)
	return newPair, nil
}

// CommonKey returns the shared key computed with the public key and the
// private key. By using Exchange, then calling CommonKey on the resulting
// KeyPair you get a key that can be used to communicate with the other side.
func (kp KeyPair) CommonKey() *[keySize]byte {
	return CommonKey(kp.pub, kp.priv)
}

// CommonKey calculates the key that is shared between the public and
// private keys given. Internally, this uses box.Precompute, performing a
// Diffie-Hellman key exchange.
func CommonKey(pub, priv *[keySize]byte) *[keySize]byte {
	var key = new([keySize]byte)
	box.Precompute(key, pub, priv)
	return key
}

const nonceSize = 24

// NewNonce returns a new Nonce initialized with a random value.
func NewNonce() (*[nonceSize]byte, error) {
	var nonce [nonceSize]byte
	_, err := io.ReadFull(rand.Reader, nonce[:])
	if err != nil {
		return nil, err
	}
	return &nonce, nil
}

// NonceFrom returns a new Nonce initialized by reading from the buffer.
// If the buffer is bigger than 24 bytes, only the first 24 bytes are read.
// An error is returned if fewer than 24 bytes are read.
func NonceFrom(buf []byte) (*[nonceSize]byte, error) {
	var n [nonceSize]byte
	c := copy(n[:], buf)
	if c < nonceSize {
		return nil, fmt.Errorf("did not write the entire value (wrote %d)", c)
	}
	return &n, nil
}
