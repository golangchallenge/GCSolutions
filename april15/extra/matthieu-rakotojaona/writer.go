//  gochallenge2 - Solution for 2nd go challenge
//
//  Written in 2015 by Matthieu Rakotojaona - matthieu.rakotojaona@gmail.com
//
//  To the extent possible under law, the author(s) have dedicated all
//  copyright and related and neighboring rights to this software to the
//  public domain worldwide. This software is distributed without any
//  warranty.
//
//  You should have received a copy of the CC0 Public Domain Dedication
//  along with this software. If not, see
//  <http://creativecommons.org/publicdomain/zero/1.0/>.

package main

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"

	"golang.org/x/crypto/nacl/box"
)

var (
	ErrTooBig           error = errors.New("Input too big (> 32kiB)")
	ErrShortSecureWrite error = errors.New("Short secure write")
)

type secureWriter struct {
	priv *[32]byte
	pub  *[32]byte
	to   io.Writer
}

// Write writes the content of p encrypted with the priv and pub
// fields of sw into the writer in sw.to, and returns the number of
// cleartext bytes written to it.
//
// If p is bigger than 32kiB, Write fails with ErrTooBig.
func (sw secureWriter) Write(p []byte) (n int, err error) {
	// Reject if message is too long
	if len(p) > 32*1024 {
		return 0, ErrTooBig
	}

	// Prepare nonce
	var nonce [24]byte
	nNonce, err := rand.Read(nonce[:])
	if err != nil {
		return 0, err
	}
	if nNonce != 24 {
		return 0, ErrShortSecureWrite
	}

	// Prepare message
	message := make([]byte, 2+len(nonce)+box.Overhead+len(p))
	contentLength := len(nonce) + box.Overhead + len(p)
	binary.BigEndian.PutUint16(message, uint16(contentLength))
	copy(message[2:], nonce[:])

	// alias for the part where we write tag and ciphertext
	out := message[len(nonce)+2:]

	// Note: we could precompute the shared key and re-use it. Not doing
	// it makes for a simpler design, only a real-world benchmark would
	// tell us if we actually need it.
	box.Seal(out[:0], p, &nonce, sw.pub, sw.priv)

	nFull, err := sw.to.Write(message)
	if err != nil {
		return 0, err
	}
	if nFull != len(message) {
		return 0, ErrShortSecureWrite
	}

	return len(p), nil
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	return secureWriter{
		priv: priv,
		pub:  pub,
		to:   w,
	}
}
