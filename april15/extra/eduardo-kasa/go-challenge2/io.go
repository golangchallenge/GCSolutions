package gc2

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"crypto/rand"

	"code.google.com/p/go.crypto/nacl/box"
)

// SecureReader implements a secure io.Reader object.
type SecureReader struct {
	sKey [32]byte
	rd   io.Reader
	zr   io.Reader
	buf  *bytes.Buffer
}

// Read reads data into p.
// Returns the number of bytes read into p.
func (sr *SecureReader) Read(p []byte) (n int, err error) {
	n = len(p)
	if n == 0 {
		return 0, nil
	}
	// check for empty buffer
	if sr.buf.Len() == 0 {
		// read message size
		var msgLen uint16
		err = binary.Read(sr.rd, binary.LittleEndian, &msgLen)
		if err == io.EOF {
			return 0, err
		} else if err != nil {
			return 0, fmt.Errorf("problem reading msg size: %s", err)
		}
		// read encrypted contento into msg
		msg := make([]byte, msgLen)
		_, err = io.ReadFull(sr.rd, msg)
		if err == io.EOF {
			return 0, err
		} else if err != nil {
			return 0, fmt.Errorf("problem reading msg size: %s", err)
		}
		dm, err := sr.decrypt(msg)
		if err != nil {
			return 0, err
		}
		// copy some decrypted content to p
		n = copy(p, dm)
		if n == len(dm) {
			// all content of dm is in p, return
			return n, nil
		}
		// p is full and dm[n:] is the remaining data
		sr.buf.Write(dm[n:])
		return n, nil
	}
	return sr.buf.Read(p)
}

// decrypt decrypts a message using a precomputed sharedKey.
func (sr *SecureReader) decrypt(message []byte) ([]byte, error) {
	var nonce [24]byte
	copy(nonce[:], message)
	opened, ok := box.OpenAfterPrecomputation(nil, message[24:], &nonce, &sr.sKey)
	if !ok {
		return nil, fmt.Errorf("decrypt: error opening box")
	}
	return opened, nil
}

// NewSecureReader returns a new SecureReader.
// This io.Reader uses peersPublicKey and privateKey to decrypt any data read from r.
func NewSecureReader(r io.Reader, privateKey, peersPublicKey *[32]byte) io.Reader {
	sr := &SecureReader{rd: r, buf: new(bytes.Buffer)}
	box.Precompute(&sr.sKey, peersPublicKey, privateKey)
	return sr
}

// SecureWriter implements a secure io.Writer object.
type SecureWriter struct {
	sKey [32]byte
	wr   io.Writer
	zw   io.Writer
}

// Write encrypts the contents of p and write into the underlying data stream.
func (sw *SecureWriter) Write(p []byte) (n int, err error) {
	box := sw.encrypt(p)
	msgLen := uint16(len(box))
	err = binary.Write(sw.wr, binary.LittleEndian, msgLen)
	if err != nil {
		return 0, err
	}
	_, err = sw.wr.Write(box)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

// Encrypt encrypts a msg using precomputed key.
// Returns [nonce|encrypted msg]
// nonce is a 24-byte array
func (sw *SecureWriter) encrypt(message []byte) []byte {
	var nonce [24]byte
	rand.Read(nonce[:])
	return box.SealAfterPrecomputation(nonce[:], message, &nonce, &sw.sKey)
}

// NewSecureWriter returns a new SecureWriter.
// This io.Writer uses peersPublicKey and privateKey to encrypt any data written into w.
func NewSecureWriter(w io.Writer, privateKey, peersPublicKey *[32]byte) io.Writer {
	sw := &SecureWriter{wr: w}
	box.Precompute(&sw.sKey, peersPublicKey, privateKey)
	return sw
}
