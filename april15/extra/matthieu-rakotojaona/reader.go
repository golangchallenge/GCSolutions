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
	"bufio"
	"encoding/binary"
	"errors"
	"io"

	"golang.org/x/crypto/nacl/box"
)

var (
	ErrShortSecureRead error = errors.New("Short secure read")
	ErrBadMessage      error = errors.New("Invalid signature")
)

type secureReader struct {
	priv *[32]byte
	pub  *[32]byte
	from io.Reader
}

// Read reads a complete length-prefixed message from sr.from and
// decrypts it into p. It returns the number of bytes read from thea
// reader.
//
// If the message has a wrong signature, it returns
// an ErrBadMessage.
func (sr secureReader) Read(p []byte) (n int, err error) {
	var l uint16
	err = binary.Read(sr.from, binary.BigEndian, &l)
	if err != nil {
		return 0, err
	}

	content := make([]byte, l)
	nn, err := sr.from.Read(content)
	if err != nil {
		return 0, err
	}
	if nn != len(content) {
		return 0, ErrShortSecureRead
	}

	// Make sure p is big enough
	if len(p) < int(l)-24-box.Overhead {
		return 0, ErrShortSecureRead
	}

	var nonce [24]byte
	copy(nonce[:], content[:])

	_, ok := box.Open(p[:0], content[len(nonce):], &nonce, sr.pub, sr.priv)
	if !ok {
		err = ErrBadMessage
	}

	return int(l) - len(nonce) - box.Overhead, err
}

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	return secureReader{
		priv: priv,
		pub:  pub,

		// A buffered reader allows us to not pay the full price for the
		// multiple reads we will do in Read()
		from: bufio.NewReader(r),
	}
}
