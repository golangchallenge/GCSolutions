// Package main contains functions and types that provide a secure
// communication stream.
//
// The protocol begins with a handshake between the two endpoints. The
// handshake is for each endpoint to send a 32 byte public key to the other
// endpoint. After that, the communication is broken into messages. Each
// message starts with a 24 byte nonce (which should be randomly generated).
// The next 2 bytes encodes the number of bytes remaining in the message. The
// remaining part of the message is encoded with the
// golang.org/x/crypto/nacl/box package.
//
//
package main

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"

	"golang.org/x/crypto/nacl/box"
)

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	return &secureReader{priv: priv, pub: pub, reader: r}
}

type secureReader struct {
	priv, pub *[32]byte
	reader    io.Reader
}

func (r *secureReader) Read(p []byte) (int, error) {

	// Read the nonce -- 24 bytes.
	var nonce [24]byte
	if _, err := io.ReadFull(r.reader, nonce[:]); err != nil {
		return 0, err
	}

	// Read the message size -- 2 bytes.
	var size uint16
	if err := binary.Read(r.reader, binary.LittleEndian, &size); err != nil {
		return 0, err
	}

	// Read in the message body -- variable bytes.
	body := make([]byte, size)
	if _, err := io.ReadFull(r.reader, body); err != nil {
		return 0, err
	}

	// Decrypt the message body.
	var decrypted []byte
	decrypted, ok := box.Open(nil, body, &nonce, r.pub, r.priv)
	if !ok {
		return 0, errors.New("could not open box")
	}
	n := copy(p, decrypted)
	return n, nil
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	return &secureWriter{priv: priv, pub: pub, writer: w}
}

type secureWriter struct {
	priv, pub *[32]byte
	writer    io.Writer
}

func (w *secureWriter) Write(p []byte) (int, error) {

	// Generate and write the nonce.
	var nonce [24]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return 0, err
	}
	if _, err := w.writer.Write(nonce[:]); err != nil {
		return 0, err
	}

	// Write the encoded message size.
	size := len(p) + box.Overhead
	if err := binary.Write(w.writer, binary.LittleEndian, uint16(size)); err != nil {
		return 0, err
	}

	// Write the body.
	encrypted := box.Seal(nil, p, &nonce, w.pub, w.priv)
	if _, err := w.writer.Write(encrypted); err != nil {
		return 0, err
	}
	return len(p), nil
}

// secureConnection creates a secure connection from an insecure connection. It
// does this by performing the handshake, then returning an io.ReadWriteCloser
// that can be used to communicate securely.
func secureConnection(conn io.ReadWriteCloser) (io.ReadWriteCloser, error) {

	var priv, pub *[32]byte
	var err error
	pub, priv, err = box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	if _, err := conn.Write(pub[:]); err != nil {
		return nil, err
	}

	if _, err := io.ReadFull(conn, pub[:]); err != nil {
		return nil, err
	}

	return struct {
		io.Reader
		io.Writer
		io.Closer
	}{
		NewSecureReader(conn, priv, pub),
		NewSecureWriter(conn, priv, pub),
		conn,
	}, nil
}
