// This file contains the Reader and Writer that perform the encryption and decryption
// using the nacl package.
package main

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"
	"net"

	nacl "golang.org/x/crypto/nacl/box"
)

const (
	// bufferSize is the size in bytes of buffers used by Read/Write to avoid extra allocations.
	bufferSize = 64 * 1024

	// nonceSize is the size of the nonce in bytes.
	nonceSize = 24

	// noncePrefix is the number of bytes used for a random per-connection prefix. The last 8 bytes
	// are used as a uint64 counter incremented on each message.
	noncePrefix = nonceSize - 8
)

// ErrInvalidData is returned if Read is unable to decrypt the data.
var ErrInvalidData = errors.New("invalid encrypted data")

// secureReader is an io.Reader that decrypts data sent using a secureWriter.
type secureReader struct {
	sharedKey  [32]byte
	underlying io.Reader

	packet *dataPacket
	buffer []byte

	// decrypted stores unconsumed decrypted bytes from the last read packet.
	decrypted []byte
}

// readDecoded copies unconsumed bytes from a previously decrypted packet into bs.
func (s *secureReader) readDecoded(bs []byte) int {
	copied := copy(bs, s.decrypted)
	s.decrypted = s.decrypted[copied:]
	return copied
}

func (s *secureReader) Read(bs []byte) (int, error) {
	if len(s.decrypted) > 0 {
		return s.readDecoded(bs), nil
	}

	// Read the next packet from the underlying reader.
	p := s.packet
	err := p.Read(s.underlying, s.buffer)
	if err != nil {
		return 0, err
	}

	// Decode the packet into bs where possible, bs[:0] will return a slice with 0 length
	// but reuse the underlying array's capacity when appending, avoiding memory allocations.
	decrypted, valid := nacl.OpenAfterPrecomputation(bs[:0], p.data, &p.nonce, &s.sharedKey)
	if !valid {
		return 0, ErrInvalidData
	}

	// If the decrypted data was smaller than the size of bs, then bs contains all the data.
	read := len(p.data) - nacl.Overhead
	if read <= len(bs) {
		s.decrypted = nil
		return read, nil
	}

	// Otherwise, the data was too large for bs, and needs to be copied into bs.
	s.decrypted = decrypted
	return s.readDecoded(bs), nil
}

// NewSecureReader returns a secureReader initialied with the given parameters.
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	s := &secureReader{
		underlying: r,
		packet:     &dataPacket{},
		buffer:     make([]byte, bufferSize),
	}
	nacl.Precompute(&s.sharedKey, pub, priv)
	return s
}

// secureWriter is an io.Writer that encrypts data before writing it to the underlying writer.
type secureWriter struct {
	sharedKey  [32]byte
	underlying io.Writer

	initErr error

	packet   *dataPacket
	buffer   []byte
	msgNonce uint64
}

func (s *secureWriter) Write(bs []byte) (int, error) {
	if s.initErr != nil {
		return 0, s.initErr
	}

	// Increment the message counter part of the nonce.
	s.msgNonce++
	binary.LittleEndian.PutUint64(s.packet.nonce[noncePrefix:], s.msgNonce)

	// Write out the data (using buffer to try and avoid extra memory allocations).
	s.packet.data = nacl.SealAfterPrecomputation(s.buffer, bs, &s.packet.nonce, &s.sharedKey)
	n, err := s.packet.Write(s.underlying)

	// Number of bytes written should be 0 <= n <= len(bs). Data written to the underlying writer
	// is larger due to Overhead, so subtract Overhead from the returned n.
	if n >= nacl.Overhead {
		n -= nacl.Overhead
	}
	return n, err
}

// NewSecureWriter returns a secureWriter initialied with the given parameters.
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	s := &secureWriter{
		underlying: w,
		packet:     &dataPacket{},
		buffer:     make([]byte, 0, bufferSize),
	}
	nacl.Precompute(&s.sharedKey, pub, priv)

	// Generate a random prefix for the nonce for this connection.
	if _, err := io.ReadFull(rand.Reader, s.packet.nonce[:noncePrefix]); err != nil {
		s.initErr = err
	}
	return s
}

type secureReadWriter struct {
	io.Reader
	io.Writer
	c net.Conn
}

// Close just closes the underlying connection.
func (s *secureReadWriter) Close() error {
	return s.c.Close()
}

func newSecureReadWriter(conn net.Conn) (io.ReadWriteCloser, error) {
	// Generate the public/private keys for this client.
	pub, priv, err := nacl.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	// Send our public key to the peer.
	if _, err := conn.Write(pub[:]); err != nil {
		return nil, err
	}

	// Receive the peer's public key.
	var peerPub [32]byte
	if _, err := io.ReadFull(conn, peerPub[:]); err != nil {
		return nil, err
	}

	return &secureReadWriter{
		NewSecureReader(conn, priv, &peerPub),
		NewSecureWriter(conn, priv, &peerPub),
		conn,
	}, nil
}
