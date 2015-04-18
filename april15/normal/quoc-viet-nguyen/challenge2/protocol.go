package main

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"unsafe"

	"golang.org/x/crypto/nacl/box"
)

/*
Each message is encrypted in a frame which contains 32-byte header. The header
includes a nonce which is generated for each message and length of the
following encrypted message.

  |<---  Header  --->|<-----  Body  ---->|
  +---------+--------+-------------------+
  |  Nonce  | Length | Encrypted Message |
  | 24-byte | 4-byte |      n-byte       |
  +---------+--------+-------------------+

In order to prevent replay, the sequence is kept synchronized between reader
and writer. Each writing will increase this sequence and the reader should
reject if the received sequence is not larger than the number in the last
verified packet (http://cr.yp.to/highspeed/coolnacl-20120725.pdf - page 5).
*/
const (
	noncePos   = 0
	nonceSize  = 24
	lengthPos  = noncePos + nonceSize
	lengthSize = 4
	headerSize = lengthPos + lengthSize

	maxMessageSize = 32 * 1024 // 32KB

	keySize = 32 // Size of public/private key
)

var (
	// ErrInvalidHeader is returned when header in a received frame is
	// invalid, i.e. length is too small or too big or packet is out
	// out sequence.
	ErrInvalidHeader = errors.New("invalid header")
	// ErrInvalidMessage indicates the reader is unable to decrypt the
	// message from a received frame, i.e. encrypted data is contaminated.
	ErrInvalidMessage = errors.New("invalid message")
)

// secureChannel is the common data structure for transferring messages
// securely.
type secureChannel struct {
	sharedKey [keySize]byte
	sequence  uint64
}

// generateNonce creates a nonce randomly. The current sequence number is also
// included in the last part of the nonce.
func (s *secureChannel) generateNonce(nonce *[nonceSize]byte) error {
	if _, err := rand.Read(nonce[:nonceSize-8]); err != nil {
		return err
	}
	binary.BigEndian.PutUint64(nonce[nonceSize-8:], s.sequence)
	return nil
}

// secureWriter encrypts messages and writes to underlying writer.
type secureWriter struct {
	writer io.Writer

	secureChannel
}

// NewSecureWriter instantiates a new SecureWriter.
func NewSecureWriter(w io.Writer, priv, pub *[keySize]byte) io.Writer {
	s := &secureWriter{
		writer: w,
	}
	box.Precompute(&s.sharedKey, pub, priv)
	return s
}

// Write encrypts message and sends to underlying writer.
// Write is not concurrent-safe due to keeping sequence for each message.
func (s *secureWriter) Write(p []byte) (int, error) {
	if len(p) == 0 || len(p) > maxMessageSize {
		return 0, fmt.Errorf("message is empty or too big: %d (max: %d)",
			len(p), maxMessageSize)
	}
	// Create a frame buffer for sending encrypted message.
	frameSize := headerSize + len(p) + box.Overhead
	frame := make([]byte, headerSize, frameSize)
	// Variable nonce is a pointer to a part of the frame buffer.
	// A new nonce is generated for each message.
	nonce := (*[nonceSize]byte)(unsafe.Pointer(&frame[noncePos]))
	s.sequence++
	if err := s.generateNonce(nonce); err != nil {
		return 0, fmt.Errorf("could not generate nonce: %v", err)
	}
	// Put length of encrypted message after nonce in the header.
	binary.BigEndian.PutUint32(frame[lengthPos:], uint32(len(p)+box.Overhead))

	// Fill encrypted message which will be box.Overhead bytes longer than
	// the original.
	_ = box.SealAfterPrecomputation(frame[headerSize:], p, nonce, &s.sharedKey)
	n, err := s.writer.Write(frame[:frameSize])
	if err != nil {
		return 0, err
	}
	if n != frameSize {
		return 0, io.ErrShortWrite
	}
	return len(p), nil
}

// secureReader reads from underlying reader and decrypts messages.
type secureReader struct {
	reader io.Reader

	secureChannel
}

// NewSecureReader instantiates a new SecureReader.
func NewSecureReader(r io.Reader, priv, pub *[keySize]byte) io.Reader {
	s := &secureReader{
		reader: r,
	}
	box.Precompute(&s.sharedKey, pub, priv)
	return s
}

// Read decrypts message from underlying reader.
// Data in underlying reader should be flushed if an error occurs.
// Read is not concurrent-safe.
func (s *secureReader) Read(p []byte) (int, error) {
	var header [headerSize]byte
	// Read header first.
	if _, err := io.ReadFull(s.reader, header[:]); err != nil {
		return 0, err
	}
	nonce := (*[nonceSize]byte)(unsafe.Pointer(&header[noncePos]))
	length := int(binary.BigEndian.Uint32(header[lengthPos:]))
	// Validate header.
	if length < box.Overhead || length > (maxMessageSize+box.Overhead) {
		// Invalid length.
		return 0, ErrInvalidHeader
	}
	sequence := binary.BigEndian.Uint64(nonce[nonceSize-8:])
	if sequence <= s.sequence {
		// Invalid sequence
		return 0, ErrInvalidHeader
	}

	// Read encrypted message.
	encrypted := make([]byte, length)
	if _, err := io.ReadFull(s.reader, encrypted); err != nil {
		return 0, err
	}
	// Decrypt the message. The output will be box.Overhead bytes smaller
	// than encrypted message.
	messageLength := length - box.Overhead
	if len(p) < messageLength {
		return 0, io.ErrShortBuffer
	}
	_, ok := box.OpenAfterPrecomputation(p[:0], encrypted, nonce, &s.sharedKey)
	if !ok {
		return 0, ErrInvalidMessage
	}
	s.sequence = sequence
	return messageLength, nil
}

// generateKeyPair creates a pair of public and private key. Since
// NewSecureReader and NewSecureWriter take priv-pub order as in parameters,
// this function returns priv then pub, not pub-priv like box.GenerateKey.
func generateKeyPair() (priv, pub *[keySize]byte, err error) {
	pub, priv, err = box.GenerateKey(rand.Reader)
	return
}
