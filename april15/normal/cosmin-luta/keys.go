//
// Go Challenge 2
//
// Cosmin Luță <q4break@gmail.com>
//
// keys.go - Key generation
//

package main

import (
	"crypto/rand"

	"golang.org/x/crypto/nacl/box"
)

// A KeyPair is a structure which contains a <public,private> key pair.
type KeyPair struct {
	pub, priv *[KeySize]byte
}

// NewKeyPair generates and returns a new KeyPair.
func NewKeyPair() (*KeyPair, error) {

	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	return &KeyPair{
		pub:  pub,
		priv: priv,
	}, nil
}
