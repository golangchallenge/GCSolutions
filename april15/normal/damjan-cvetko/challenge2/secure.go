// This file implements the SecureReader, SecureWriter and SecureReadWriteCloser.
// These wrap golang.org/x/crypto/nacl/box boxing methods.
//
// The wire protocol constists of a fixed length nonce,
// a encrypted lenght of payload in 16bit signed network byte order (BigEndian)
// and then the variable lenght encrypted/sealed payload
package main

import (
	"bufio"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"

	"golang.org/x/crypto/nacl/box"
)

// ErrFailedCreatingNonce means that Write operation could not generate a new
// nonce
var ErrFailedCreatingNonce = errors.New("secure: Failed to create nonce")

// ErrMessageNotValid means that the read boxed message could not be verified
// against private key
var ErrMessageNotValid = errors.New("secure: Message not valid")

const (
	maxMessageLen = 100 //32768
)

// SecureReader implements reading of NaCL boxed messages from the underlying
// io.Reader. It uses box.Open method for decrytion.
// Each underlying message consists of a unencrypted header (msgHead) and
// encrypted payload.
// Read method will read a whole message from the wrie, decrypt it and store it
// in an internal buffer. It will then return as much of the message as possible.
// If the Read buffer is too small, the rest of the decrypted message will be
// returned on subsequent calls. If the Read buffer is larger than the message,
// only the first message will be returned. Next message will be read from
// the wire on next call to Read.
// All buffers are part of the SecureReader struct and no extra allocations
// should be needed during reading
type SecureReader struct {
	r         io.Reader
	sharedKey [32]byte
	boxBuf    [maxMessageLen + box.Overhead]byte
	buf       [maxMessageLen]byte
	unreadBuf []byte
}

// Reads and decrypts a message from the underlying reader, returns as much as
// possible from the read message into p. Further reads return the rest of the
// message. Only when one message is compeltely returned to caller, the next
// message will be read from the wire.
func (sr *SecureReader) Read(p []byte) (n int, err error) {
	if sr.unreadBuf == nil || len(sr.unreadBuf) == 0 {
		h := msgHead{}
		if err = binary.Read(sr.r, binary.BigEndian, &h); err != nil {
			return 0, err
		}
		if _, err = io.ReadFull(sr.r, sr.boxBuf[:h.Length]); err != nil {
			return 0, err
		}
		var ok bool
		if sr.unreadBuf, ok = box.OpenAfterPrecomputation(sr.buf[:0], sr.boxBuf[:h.Length], &h.Nonce, &sr.sharedKey); !ok {
			return 0, ErrMessageNotValid
		}
	}
	n = copy(p, sr.unreadBuf)
	sr.unreadBuf = sr.unreadBuf[n:]
	return n, nil
}

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, privateKey, peersPublicKey *[32]byte) io.Reader {
	ret := &SecureReader{
		r: r,
	}
	ret.unreadBuf = ret.buf[:0]
	box.Precompute(&ret.sharedKey, peersPublicKey, privateKey)
	return ret
}

// SecureWriter implements writing of NaCL boxed messages to the underlying
// io.Writer. It uses box.Seal method for encryption.
// Each message on the wire consists of a unencrypted header (msgHead) and
// encrypted payload.
type SecureWriter struct {
	w         io.Writer
	sharedKey [32]byte
	boxBuf    [maxMessageLen + box.Overhead]byte
}

// Write encrypts the contents of the provided buffer and sends it along with
// the required header. Each Write results in one message on the wire.
func (sw *SecureWriter) Write(p []byte) (n int, err error) {
	buf := sw.boxBuf[:len(p)+box.Overhead]
	h, err := newHead(len(buf))
	if err != nil {
		return 0, ErrFailedCreatingNonce
	}
	// bufio is used so that all data is sent in one packet
	bw := bufio.NewWriterSize(sw.w, binary.Size(h)+len(buf))

	if err = binary.Write(bw, binary.BigEndian, h); err != nil {
		// if there is an error at this point, don't waste time encrypting
		return 0, err
	}
	box.SealAfterPrecomputation(buf[:0], p, &h.Nonce, &sw.sharedKey)

	bw.Write(buf)
	return len(p), bw.Flush()
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, privateKey, peersPublicKey *[32]byte) io.Writer {
	ret := &SecureWriter{w: w}
	box.Precompute(&ret.sharedKey, peersPublicKey, privateKey)
	return ret
}

// SecureReadWriteCloser combines naclReader and naclWriter and adds the Close
// function. It can wrap a io.ReadWriteCloser. Usefull for wrapping a net.Conn.
type SecureReadWriteCloser struct {
	*SecureReader
	*SecureWriter
	closer io.Closer
}

// Close passes Close call to underlying io.Closer
func (naclrwc *SecureReadWriteCloser) Close() error {
	return naclrwc.closer.Close()
}

// NewSecureReadWriteCloser instantiates a new SecureReadWriteCloser
func NewSecureReadWriteCloser(rwc io.ReadWriteCloser, privateKey, peersPublicKey *[32]byte) io.ReadWriteCloser {
	ret := &SecureReadWriteCloser{
		closer: rwc,
	}
	ret.SecureReader = NewSecureReader(rwc, privateKey, peersPublicKey).(*SecureReader)
	ret.SecureWriter = NewSecureWriter(rwc, privateKey, peersPublicKey).(*SecureWriter)
	return ret
}

// msgHead is the structure that is added before each encrypted payload
// It consists of a nonce that is unique for each sent message and a
// message length.
// Both fields are sent over the wire unencrypted
//
// Note: As NaCL box does not attempt to hide the lenght of the payload
// and we are inteding to use these wrappers with TCP transport where
// (assuming single message is sent to the wire at one time) length
// is easily seen, this should not be a problem.
// Length is needed if we want to be able to use this wrapper for more than
// one message
//
// Note2: int16 was chosen for Length as challenge notes that max message is 32KB
type msgHead struct {
	Nonce  [24]byte
	Length uint16
}

// newHead creates a new msgHead structure and generates a random nonce
// while also storing the provided length for message
func newHead(length int) (msgHead, error) {
	h := msgHead{
		Length: uint16(length),
	}
	if _, err := rand.Read(h.Nonce[:]); err != nil {
		return msgHead{}, err
	}
	return h, nil
}
