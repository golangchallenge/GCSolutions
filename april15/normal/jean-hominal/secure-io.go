package main

import (
	"crypto/rand"
	"errors"
	"io"

	"golang.org/x/crypto/nacl/box"
)

const blockSize = 0x100

type secureIo struct {
	sharedKey [32]byte
	nonce     [24]byte
	// The first error returned by the underlying io object
	err error
	// An array that contains the crypted data to avoid allocating
	encryptedBlockArray [blockSize + box.Overhead + 2]byte
	// An array that contains the non-encrypted data to avoid allocating
	clearBlockArray [blockSize + 1]byte
}

func (s *secureIo) incrementNonce() {
	for i := 0; i < len(s.nonce); i++ {
		s.nonce[i]++
		if s.nonce[i] != 0 {
			break
		}
	}
}

func newSecureIo(priv, pub *[32]byte) secureIo {
	var sharedKey [32]byte
	box.Precompute(&sharedKey, pub, priv)

	return secureIo{sharedKey: sharedKey}
}

type decryptReader struct {
	secureIo
	// The underlying reader, from which encrypted bytes are read
	io.Reader
	// Whether the nonce has been read
	nonceRead bool
	// The bytes that are already decrypted, but have not been read yet
	unreadBytes      []byte
	clearBlockLength [1]byte
}

func (r *decryptReader) Read(p []byte) (n int, err error) {
	if r.err != nil {
		err = r.err
		return
	}

	// Read nonce from 24 first bytes.
	if !r.nonceRead {
		_, err = io.ReadFull(r.Reader, r.nonce[:])
		r.nonceRead = true
	}

	// If unread bytes have already been decrypted
	// then return these bytes.
	if len(r.unreadBytes) != 0 {
		n = r.readUnreadBytes(p)
		return
	}

	// No bytes left unread - will have to decrypt the next block.

	// First, read the first byte for the length of clear data.
	// err may be non-nil if there was an error while reading the nonce.
	if err == nil {
		_, err = io.ReadFull(r.Reader, r.clearBlockLength[:])
	}

	// Then allocate a slice of the correct size to read the next encrypted block
	eBlock := []byte(nil)
	if err == nil {
		l := int(r.clearBlockLength[0]) + 1
		eBlock = r.encryptedBlockArray[:l+box.Overhead+1]

		_, err = io.ReadFull(r.Reader, eBlock)
	}

	// Decrypt the block and validate its size
	cBlock := []byte(nil)
	if (err == nil || err == io.EOF) && len(eBlock) != 0 {
		cBlock = r.clearBlockArray[:len(eBlock)-box.Overhead]

		cBlock, valid := box.OpenAfterPrecomputation(cBlock[:0], eBlock, &r.nonce, &r.sharedKey)
		r.incrementNonce()

		if !valid || int(cBlock[0])+2 != len(cBlock) {
			err = errors.New("The reader is reading corrupted data.")
		}
	}

	// Set unreadBytes, and return as many unread bytes as possible
	if (err == nil || err == io.EOF) && len(cBlock) != 0 {
		r.unreadBytes = cBlock[1:]
		n = r.readUnreadBytes(p)
	}

	r.err = err
	return
}

func (r *decryptReader) readUnreadBytes(p []byte) (n int) {
	n = len(r.unreadBytes)
	if len(p) < n {
		n = len(p)
	}
	copy(p[:n], r.unreadBytes[:n])
	r.unreadBytes = r.unreadBytes[n:]
	return
}

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	if r == nil || priv == nil || pub == nil {
		return nil
	}

	sio := newSecureIo(priv, pub)

	return &decryptReader{secureIo: sio, Reader: r}
}

type encryptWriter struct {
	secureIo
	// The underlying writer, where encrypted bytes should be written to
	io.Writer
	// Whether the nonce has been written
	nonceWritten bool
}

func (w *encryptWriter) Write(p []byte) (n int, err error) {
	if w.err != nil {
		return 0, w.err
	}

	// Write generated nonce so that reader can find it
	if !w.nonceWritten {
		_, w.err = w.Writer.Write(w.nonce[:])
		w.nonceWritten = true
	}

	// Write blocks of length 256 at most.
	i := 0
	for i = 0; i+blockSize-1 < len(p) && w.err == nil; i += blockSize {
		n += w.writeBlock(p[i : i+blockSize])
	}

	// Final block, which may be shorter than 255
	if i < len(p) && w.err == nil {
		n += w.writeBlock(p[i:])
	}

	err = w.err
	return
}

func (w *encryptWriter) writeBlock(b []byte) int {
	// Length prefix before the encrypted box so that the
	// reader knows how much data should be read;
	eBlock := w.encryptedBlockArray[:len(b)+box.Overhead+2]
	eBlock[0] = byte(len(b) - 1)

	// Length prefix inside the encrypted box so that the
	// reader will be able to check that the unencrypted
	// length was not tampered with;
	cBlock := w.clearBlockArray[:len(b)+1]
	cBlock[0] = byte(len(b) - 1)

	copy(cBlock[1:], b)

	// Encrypt the data.
	eBlock = box.SealAfterPrecomputation(eBlock[:1], cBlock, &w.nonce, &w.sharedKey)
	w.incrementNonce()

	_, w.err = w.Writer.Write(eBlock)

	if w.err != nil {
		return 0
	}

	return len(b)
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	if w == nil || priv == nil || pub == nil {
		return nil
	}

	// Generate random nonce
	sio := newSecureIo(priv, pub)
	_, err := rand.Read(sio.nonce[:])

	if err != nil {
		return nil
	}

	return &encryptWriter{secureIo: sio, Writer: w}
}
