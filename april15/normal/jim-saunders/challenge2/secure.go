package main

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"golang.org/x/crypto/nacl/box"
	"io"
	"log"
)

// SecureMessage is the message format the secure.Reader
// and secure.Writer use to communicate
type SecureMessage struct {
	Nonce [24]byte
	Body  []byte
}

// SecureReader reads a SecureMessage and decrypts the
// body field of the message
type SecureReader struct {
	io.Reader
	Private *[32]byte
	Public  *[32]byte
}

//SecureReader.Read reads a secure.Message and
//unencrypts the body portion of the message.
func (s SecureReader) Read(p []byte) (int, error) {
	var msg SecureMessage
	var out []byte

	d := json.NewDecoder(s.Reader)
	err := d.Decode(&msg)

	if err == nil {
		if b, success := box.Open(out, msg.Body, &msg.Nonce, s.Public, s.Private); !success {
			err = errors.New("Could not decrypt message!")
		} else {
			return copy(p, b), nil
		}
	}
	log.Fatal(err)
	return 0, err
}

// SecureWriter writes a SecureMessage generating a
// random nounce and encrypting the body field of the
// message before writing it out in JSON format.
type SecureWriter struct {
	io.Writer
	Private *[32]byte
	Public  *[32]byte
}

// Generate a random 24 byte array
func (s SecureWriter) nonce() [24]byte {
	var out [24]byte
	var bytes = make([]byte, len(out))
	var alnum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

	rand.Read(bytes)
	for i, b := range bytes {
		out[i] = alnum[b%byte(len(alnum))]
	}
	return out
}

// Generate a nounce and encrypt the body of a SecureMessage.
// Finally write it out in JSON format.
func (s SecureWriter) Write(p []byte) (int, error) {
	var out []byte
	non := s.nonce()
	msg := SecureMessage{non, box.Seal(out, p, &non, s.Public, s.Private)}
	b, err := json.Marshal(msg)
	if err == nil {
		return s.Writer.Write(b)
	}
	log.Fatal(err)
	return 0, err
}
