package main

import (
	"crypto/rand"
	"encoding/binary"
	"golang.org/x/crypto/nacl/box"
	"io"
	"log"
)

type SecureWriter struct {
	w    io.Writer
	priv *[32]byte
	pub  *[32]byte
}

func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	return &SecureWriter{w, priv, pub}
}

func (sw *SecureWriter) Write(message []byte) (int, error) {
	// Generate a random nonce
	nonce, err := randomNonce()
	if err != nil {
		log.Println("Error generating nonce", err)
		return 0, err
	}

	// Convert message to encrypted byte slice with nonce
	encrypted := box.Seal(nonce[:], message, nonce, sw.pub, sw.priv)
	payloadSize := len(encrypted)

	// Write payload size to buffer
	err = binary.Write(sw.w, binary.LittleEndian, uint32(payloadSize))
	if err != nil {
		log.Println("Error writing payloadSize to buffer", err)
		return 0, err
	}

	// Write encrypted message to buffer
	_, err = sw.w.Write(encrypted)
	if err != nil {
		log.Println("Error writing encrypted message to buffer", err)
		return 0, err
	}

	return len(message), nil
}

func randomNonce() (*[24]byte, error) {
	var buf [24]byte
	_, err := rand.Read(buf[:])
	if err != nil {
		log.Println("Error generating nonce:", err)
		return nil, err
	}
	return &buf, nil
}
