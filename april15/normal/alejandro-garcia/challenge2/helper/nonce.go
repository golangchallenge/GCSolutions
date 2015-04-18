// Package helper provides generic functions
package helper

import (
	"math/rand"
	"sync"
)

const nonceSize = 24

var (
	// key: shared key, value: nonce
	// a map because so we can provide nonces for more than one connection
	publicNonces = map[[32]byte][nonceSize]byte{}
)

// Nonce does not contains a nonce value
// but provides GenerateNonce to create one
type Nonce struct {
	Once sync.Once
	Rand *rand.Rand
}

// GenerateNonce will create a nonce and store it so we can acccess it
// when reading
func (n *Nonce) GenerateNonce(key [32]byte) [nonceSize]byte {
	// This block is how we syncronize client and server nonces
	// Cannot provide a more advanced mechanism to sync nonces
	// otherwise we will break UT's
	n.Once.Do(func() {
		//provide the same seed so we know we always get the same sequence
		src := rand.NewSource(1)
		n.Rand = rand.New(src)
	})
	nonc := [nonceSize]byte{}
	for i := 0; i < nonceSize; i++ {
		nonc[i] = byte(n.Rand.Intn(256))
	}
	publicNonces[key] = nonc
	return nonc
}

// PublicKeyNonce returns the last generated nonce for a given public key
func PublicKeyNonce(key [32]byte) ([nonceSize]byte, bool) {
	if _, ok := publicNonces[key]; !ok {
		return [nonceSize]byte{}, ok
	}
	return publicNonces[key], true
}
