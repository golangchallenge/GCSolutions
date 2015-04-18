package main

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/nacl/box"
)

// maxMessageSize is the greatest number of bytes that can be transmitted as a
// message with SecureWriter.
const maxMessageSize = uint64(32 * 1024)

// maxWrittenMessageSize is the greatest number of bytes that will be
// transmitted after encryption, incoming to SecureReader.
const maxWrittenMessageSize = maxMessageSize + box.Overhead + nonceSize + 8 // uint64 header

// SecureReader implements io.Reader and uses a key to decrypt messages from
// the underlying Reader. It expects the data to be in the form defined by
// SecureWriter.
type SecureReader struct {
	r   io.Reader
	key *[32]byte
}

// Read implements io.Reader. Expects that data read from the reader has been
// encrypted. If out is not big enough to hold the decrypted message, a partial
// message is written and an error is returned.
func (r SecureReader) Read(out []byte) (int, error) {
	// Read the header to find out how big the message is.
	var size uint64
	err := binary.Read(r.r, binary.BigEndian, &size)
	if err != nil {
		return 0, err
	}
	debugf("Read: %d byte message\n", size)

	if size > maxWrittenMessageSize {
		return 0, fmt.Errorf("message is too large. Got %d bytes, max: %d", size, maxWrittenMessageSize)
	}

	// This buffer holds the encrypted message.
	buf := make([]byte, size)

	// Read everything into the buffer.
	c, err := io.ReadFull(r.r, buf)
	if err != nil {
		return 0, err
	}
	debugf("Read: %d bytes\n%s\n", c, hex.Dump(buf))

	// Get the Nonce from the buffer.
	nonce, err := NonceFrom(buf)
	if err != nil {
		return 0, err
	}
	debugf("Read: nonce\n%s\n", hex.Dump(nonce[:]))

	// The message is the rest of the buffer after the nonce.
	var msg = buf[len(nonce):]

	debugf("Read: msg\n%s\n", hex.Dump(msg))

	// Decrypt the message.
	res, ok := box.OpenAfterPrecomputation(nil, msg, nonce, r.key)
	if !ok {
		return 0, errors.New("decryption failed")
	}
	debugf("Read: result\n%s\n", hex.Dump(res))

	// Copy the result for output.
	co := copy(out, res)
	if co < len(res) {
		return co, fmt.Errorf("failed to write into the output buffer. Wrote %d, needed %d", co, len(res))
	}
	return co, nil
}

// SecureWriter implements io.Writer and encrypts data with a key before
// writing to the underlying writer. The encrypted data has a uint64 header
// indicating how long the encrypted message is, followed by the encrypted
// message. The encrypted message is a 24 byte nonce followed by the message.
type SecureWriter struct {
	w   io.Writer
	key *[32]byte
}

// Write implements io.Writer.
func (w SecureWriter) Write(buf []byte) (int, error) {
	if uint64(len(buf)) > maxMessageSize {
		return 0, fmt.Errorf("input is too large. Got %d bytes, max: %d", len(buf), maxMessageSize)
	}

	// Create a nonce.
	nonce, err := NewNonce()
	if err != nil {
		return 0, err
	}
	debugf("Write: nonce\n%s\n", hex.Dump(nonce[:]))

	// Encrypt the message with the nonce prefix.
	sealed := box.SealAfterPrecomputation(nonce[:], buf, nonce, w.key)

	debugf("Write: sealed %d bytes\n%s\n", len(sealed), hex.Dump(sealed))

	// Write a fixed header indicating how long the message is.
	header := uint64(len(sealed))
	headerSize := binary.Size(header)
	if err := binary.Write(w.w, binary.BigEndian, header); err != nil {
		return headerSize, err
	}

	// Write the message.
	messageSize, err := w.w.Write(sealed)

	// Return the size of the header and the message.
	return headerSize + messageSize, err
}
