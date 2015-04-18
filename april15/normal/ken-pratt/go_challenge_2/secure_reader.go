package main

import (
	"encoding/binary"
	"golang.org/x/crypto/nacl/box"
	"io"
	"log"
)

type ReadError struct {
	Message string
}

func (e *ReadError) Error() string {
	return e.Message
}

type SecureReader struct {
	r        io.Reader
	priv     *[32]byte
	pub      *[32]byte
	leftover []byte
}

func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	return &SecureReader{r, priv, pub, nil}
}

func (sr *SecureReader) Read(out []byte) (int, error) {
	// If there isn't a leftover buffer, then it's time to read the next encrypted message
	if sr.leftover == nil {
		err := sr.ReadNextEncryptedMessage()
		if err != nil {
			return 0, err
		}
	}

	// Send as much data as possible
	var toSend []byte
	if len(sr.leftover) > len(out) {
		// We have too much data, send what we can and stash the rest in the leftover buffer
		toSend = sr.leftover[0:len(out)]
		sr.leftover = sr.leftover[len(out):]
	} else {
		// We can fit everything, so send it and set the leftover buffer to nil
		toSend = sr.leftover
		sr.leftover = nil
	}
	copy(out, toSend)
	return len(toSend), nil
}

// Blocking read until the whole encrypted message is received
// Encrypted messages are in the format:
//   message = | 4-byte little-endian uint32 for payload size | payload |
//   payload = | 24-byte nonce | encrypted message |
func (sr *SecureReader) ReadNextEncryptedMessage() error {
	// Read the payload size out of the buffer
	var payloadSize uint32
	err := binary.Read(sr.r, binary.LittleEndian, &payloadSize)
	if err != nil {
		if err != io.EOF {
			log.Println("Error reading payloadSize from buffer", err)
		}
		return err
	}

	// Read the payload
	data := make([]byte, payloadSize)
	_, err = io.ReadFull(sr.r, data)
	if err != nil {
		log.Println("Error reading payload from buffer", err)
		return err
	}

	// Unpack the nonce and encrypted message
	nonce := data[0:24]
	encrypted := data[24:]

	// Decrypt the encrypted message
	var nonceBuf [24]byte
	copy(nonceBuf[:], nonce)
	decrypted, success := box.Open(make([]byte, 0), encrypted, &nonceBuf, sr.pub, sr.priv)
	if success {
		sr.leftover = decrypted
		return nil
	} else {
		log.Println("Error decrypting message")
		return &ReadError{"Error decrypting message"}
	}
}
