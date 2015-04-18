// Package nonce provides an implementation of an abstract nonce that can
// be incremented and compared to other nonces for security validation.
package nonce

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// Nonce is the interface that can manipulate and compare nonce data.
type Nonce interface {
	Slice() []byte
	Array() *[24]byte
	Increment()
	After(other Nonce) bool
}

// NewNonce generates a completely random new nonce.
func NewNonce() Nonce {
	var n nonceArray

	max := new(big.Int).Exp(big.NewInt(2), big.NewInt(int64(len(n)*8)), nil)
	num, _ := rand.Int(rand.Reader, max)

	copy(n[:], num.Bytes())
	return &n
}

// FromBytes creates a nonce from a given slice of 24 bytes.
func FromBytes(bytes []byte) (Nonce, error) {
	if len(bytes) < 24 {
		return nil, fmt.Errorf("byte slice must be at least 24 bytes long, found %v bytes", len(bytes))
	}
	var n nonceArray
	copy(n[:], bytes[:24])
	return &n, nil
}

// nonceArray provides the underlying implementation of a Nonce interface.
type nonceArray [24]byte

// Slice returns a byte slice pointing to the nonces underlying array.
func (n *nonceArray) Slice() []byte {
	return n[:]
}

// Array returns a pointer to the nonces byte array.
func (n *nonceArray) Array() *[24]byte {
	return (*[24]byte)(n)
}

// Increment adds to the nonce's value.
func (n *nonceArray) Increment() {
	thisInt := new(big.Int).SetBytes(n.Slice())
	thisInt.Add(thisInt, big.NewInt(1))
	copy(n[:], thisInt.Bytes())
}

// After determines whether this nonce's value is larger than another nonce's
// value or not.
func (n *nonceArray) After(other Nonce) bool {
	thisInt := new(big.Int).SetBytes(n.Slice())
	otherInt := new(big.Int).SetBytes(other.Slice())

	after := thisInt.Cmp(otherInt)
	return after > 0
}
