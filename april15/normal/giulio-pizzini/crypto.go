// This file contains all the functions for encryption
// and decryption encapsulating the NaCl library
package main

import (
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/nacl/box"
)

const (
	// maxMsgLen is the max allowed length of a message in bytes.
	maxMsgLen = 32
	// nonceLen is the length of the nonce.
	nonceLen = 24
	// maxCipherLen is the max possibe lenght of a ciphertext, which
	// is composed by concatenating three parts: (1) the nonce,
	// (2) an autehntication overhead, (3) the encrypted message.
	maxCipherLen = nonceLen + box.Overhead + maxMsgLen
)

// a key is a public, shared or private key
type key *[32]byte

// newKey returns an empty key with its underlying array
// initialized; we need it because we cannot initialize
// the array directly from the pointer type key
func newKey() key {
	return &[32]byte{}
}

// generateKeys creates and returns a public, private key pair.
func generateKeys() (pub, priv key, err error) {
	return box.GenerateKey(rand.Reader)
}

// sharedKey takes a private key and a peer public key
// to generate and return a shared key
func sharedKey(priv, peerPub key) key {
	k := newKey()
	box.Precompute(k, peerPub, priv)
	return k
}

// decrypt takes an encrypted message and a key,
// and returns the decrypted message and any error.
func decrypt(c []byte, k key) (msg []byte, err error) {
	// split the nonce and the cipher
	nonce, c, err := splitNonceCipher(c)
	if err != nil {
		return nil, err
	}

	// decrypt the message
	msg, ok := box.OpenAfterPrecomputation(msg, c, nonce, k)
	if !ok {
		return nil, fmt.Errorf("Cannot decrypt, malformed message")
	}
	return msg, nil
}

// splitNonceCipher takes a cipher, splits the nonce from
// the encrypted message and returns them and any errors.
func splitNonceCipher(c []byte) (*[nonceLen]byte, []byte, error) {
	if len(c) < nonceLen {
		return nil, nil, fmt.Errorf("cipher is shorter than nonce")
	}
	nonce := [nonceLen]byte{}
	copy(nonce[:], c)
	c = c[nonceLen:]
	return &nonce, c, nil
}

// encrypt takes a plaintext message and a key, encrypts the message,
// prepend the nonce and returns the resulting ciphertext and any errors.
func encrypt(msg []byte, k key) (c []byte, err error) {
	nonce, err := randomNonce()
	if err != nil {
		return nil, err
	}
	c = box.SealAfterPrecomputation(c, msg, nonce, k)
	c = append((*nonce)[:], c...)
	return c, nil
}

// randomNonce returns a random nonce and any errors.
func randomNonce() (*[nonceLen]byte, error) {
	nonce := &[nonceLen]byte{}
	_, err := rand.Read(nonce[:])
	if err != nil {
		return nil, err
	}
	return nonce, nil
}
