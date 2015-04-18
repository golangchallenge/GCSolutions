package sio

import (
	"crypto/rand"
	"errors"

	"golang.org/x/crypto/nacl/box"
)

func pack(priv, pub *[32]byte, data []byte) (out []byte, err error) {
	// encrypt message, build output slice nonce+message
	var nonce [24]byte

	if _, err = rand.Read(nonce[:24]); err != nil {
		return nil, errors.New("pack() failed to generate nonce: " + err.Error())
	}

	payload := box.Seal(nil, data, &nonce, pub, priv)

	out = make([]byte, 0)
	out = append(out, nonce[:]...)
	out = append(out, payload...)

	return
}

func unpack(priv, pub *[32]byte, n int, data []byte) (out []byte, err error) {
	// split nonce and payload from message, decrypt payload
	var nonce [24]byte
	var res bool

	copy(nonce[:24], data[:24])
	payload := data[24:n]

	if out, res = box.Open(nil, payload, &nonce, pub, priv); !res {
		return nil, errors.New("unpack() failed to decrypt message")
	}

	return out, nil
}
