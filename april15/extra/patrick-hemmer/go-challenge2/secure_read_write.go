// SecureWriter & SecureReader handle encryption of plain text, and
// transmitting across an underlying writer/reader.
//
// The wire protocol looks like:
//
//  [0:24]   nonce
//   [0:8]   timestamp with nanosecond precision (uint64 - network byte order)
//   [8:24]  random data
//  [24:28]  cipher text length (uint32 - network byte order)
//  [28:]    cipher text

package main

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"time"

	"golang.org/x/crypto/nacl/box"
)

// SecureWriter implements the io.Writer interface to perform NaCl encryption.
type SecureWriter struct {
	writer    io.Writer
	sharedKey [32]byte
}

// NewSecureWriter creates a new SecureWriter which wraps the provided writer.
func NewSecureWriter(w io.Writer, privateKey, peerPublicKey *[32]byte) *SecureWriter {
	sw := &SecureWriter{
		writer: w,
	}
	box.Precompute(&sw.sharedKey, peerPublicKey, privateKey)
	return sw
}

// Write encrypts the given bytes and writes them to the underlying writer.
//
// Upon any write error, the writer should not be used further as the stream
// may be in an inconsistent state.
func (sw *SecureWriter) Write(p []byte) (int, error) {
	// create & write the nonce
	var nonce [24]byte
	var nonceID = nonce[:8]
	var nonceRand = nonce[8:]
	binary.BigEndian.PutUint64(nonceID, uint64(time.Now().UnixNano()))
	if _, err := rand.Read(nonceRand); err != nil {
		return 0, err
	}
	if _, err := sw.writer.Write(nonce[:]); err != nil {
		return 0, err
	}

	// write the length of the cipher text
	var ctLen = make([]byte, 4)
	binary.BigEndian.PutUint32(ctLen, uint32(len(p)+box.Overhead))
	if _, err := sw.writer.Write(ctLen); err != nil {
		return 0, err
	}

	// generate & write the cipher text
	ct := box.SealAfterPrecomputation(nil, p, &nonce, &sw.sharedKey)
	if _, err := sw.writer.Write(ct); err != nil {
		return 0, err
	}

	return len(p), nil
}

// SecureReader implements the io.Reader interface to perform NaCl decryption.
//
// When reading a message from a peer, the first 8 bytes of the nonce must
// always be greater than that of the previously received message.
// The nonce is expected to be in network byte order (big endian).
type SecureReader struct {
	reader      io.Reader
	sharedKey   [32]byte
	buffer      []byte
	lastNonceID uint64
}

// NewSecureReader creates a new SecureReader.
func NewSecureReader(r io.Reader, privateKey, peerPublicKey *[32]byte) *SecureReader {
	sr := &SecureReader{
		reader: r,
	}
	box.Precompute(&sr.sharedKey, peerPublicKey, privateKey)
	return sr
}

// Read implements io.Reader to provide the decrypted contents of the underlying
// reader.
//
// Upon any read error, the reader should not be used further as the stream may
// be in an inconsistent state.
func (sr *SecureReader) Read(p []byte) (int, error) {
	if len(sr.buffer) > 0 {
		// we've got stuff in the buffer left from a previous read
		n := copy(p, sr.buffer)
		sr.buffer = sr.buffer[n:]
		return n, nil
	}

	// get & check the nonce
	var nonce [24]byte
	if _, err := io.ReadFull(sr.reader, nonce[:]); err != nil {
		return 0, err
	}
	nonceID := binary.BigEndian.Uint64(nonce[:8])
	if nonceID <= sr.lastNonceID {
		return 0, fmt.Errorf("invalid nonce")
	}

	// get length of cipher text
	ctLenBytes := make([]byte, 4)
	if _, err := io.ReadFull(sr.reader, ctLenBytes); err != nil {
		return 0, err
	}
	ctLen := binary.BigEndian.Uint32(ctLenBytes)

	// get cipher text
	ct := make([]byte, ctLen)
	if _, err := io.ReadFull(sr.reader, ct); err != nil {
		return 0, err
	}

	// decrypt
	ptLen := int(ctLen - box.Overhead)
	var buf []byte
	if len(p) < ptLen {
		buf = make([]byte, ptLen)
	} else {
		buf = p
	}
	buf, ok := box.OpenAfterPrecomputation(buf[:0], ct, &nonce, &sr.sharedKey)
	if !ok {
		return 0, fmt.Errorf("unable to decrypt message")
	}

	sr.lastNonceID = nonceID

	if len(p) < len(buf) || &p[len(buf)-1] != &buf[len(buf)-1] {
		// p wasn't big enough, and buf was reallocated
		// Everything after the || should not be necessary, but is there just in case
		// the buf was reallocated for some other reason.
		copy(p, buf[:len(p)])
		sr.buffer = buf[len(p):]
		return len(p), nil
	}

	return len(buf), nil
}
