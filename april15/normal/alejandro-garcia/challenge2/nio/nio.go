// Package nio provides NACL IO ReaderWriterCloser struct and functionality
package nio

import (
	"errors"
	"io"

	"challenge2/helper"

	"golang.org/x/crypto/nacl/box"
)

var Nonce helper.Nonce

type NaCLReader struct {
	Reader io.Reader
	Shared *[32]byte
}

type NaCLWriter struct {
	Writer io.Writer
	Shared *[32]byte
}

type NaCLReaderWriterCloser struct {
	Rd io.Reader
	Wr io.Writer
	Cl io.Closer
}

// Read and decrypts a message
func (nr NaCLReader) Read(p []byte) (n int, err error) {
	n, err = nr.Reader.Read(p)
	if err != nil {
		return n, err
	}
	// get the last used nonce
	nonc, ok := helper.PublicKeyNonce(*nr.Shared)
	if !ok {
		return -1, errors.New("in reader, problem while getting nonce")
	}
	out := make([]byte, n)
	dp, ok := box.OpenAfterPrecomputation(out, p[:n], &nonc, nr.Shared)
	if !ok {
		return -1, errors.New("in reader, decrypting error")
	}
	n = copy(p, dp[n:])
	return n, nil
}

// Write will encrypt message and write it
func (nw NaCLWriter) Write(p []byte) (n int, err error) {
	// generate a random nonce
	nonc := (&Nonce).GenerateNonce(*nw.Shared)
	out := make([]byte, n)
	b := box.SealAfterPrecomputation(out, p, &nonc, nw.Shared)
	return nw.Writer.Write(b)
}

func (rwc NaCLReaderWriterCloser) Read(p []byte) (n int, err error) {
	return rwc.Rd.Read(p)
}

func (rwc NaCLReaderWriterCloser) Write(p []byte) (n int, err error) {
	return rwc.Wr.Write(p)
}

func (rwc NaCLReaderWriterCloser) Close() (err error) {
	return rwc.Cl.Close()
}
