package main

import (
	"io"

	"golang.org/x/crypto/nacl/box"
)

const (
	// MaxMsgLen is the maximum length of message that can be
	// transferred over a SecureWriter/SecureReader
	MaxMsgLen = 32 * 1024

	// sizeLen is the number of bytes transferred for the
	// encrypted data lenght
	sizeLen = 4

	// nonceLen is the number of bytes transferred for the nonce
	nonceLen = 24

	// headerLen is the number of bytes transferred for all the
	// header data
	headerLen = sizeLen + nonceLen

	// MsgOverhead is the additional number of bytes transferred to
	// accommodate encryption and protocol headers
	MsgOverhead = box.Overhead + headerLen
)

// writeFull takes a payload p and writes it in full to the io.Writer w,
// continuing if a write operation is interrupted and stopping if any
// other error is reported. It always returns the total number of bytes
// that were written to the io.Writer.
func writeFull(w io.Writer, p []byte) (int, error) {
	var i int

	for i = 0; i < len(p); {
		switch n, err := w.Write(p[i:]); err {
		case nil, io.ErrShortWrite:
			i += n
		default:
			return i, err
		}
	}

	return i, nil
}

type CatchErrorReadWriter struct {
	rw  io.ReadWriter
	err error
}

func (rw *CatchErrorReadWriter) Read(p []byte) (int, error) {
	n := 0

	if rw.err == nil {
		n, rw.err = rw.rw.Read(p)
	}

	return n, rw.err
}

func (rw *CatchErrorReadWriter) Write(p []byte) (int, error) {
	n := 0

	if rw.err == nil {
		n, rw.err = rw.rw.Write(p)
	}

	return n, rw.err
}
