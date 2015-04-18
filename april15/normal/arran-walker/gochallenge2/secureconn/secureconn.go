// Package secureconn implements secured I/O by wrapping and implementing the
// standard io.Reader and io.Writer interfaces.
//
// Data written with SecureWriter and read with SecureReader encrypt and decrypt
// data to NaCl boxes respectively.
//
// A higher level implementation (SecureConn) wraps and provides a
// ReadWriteCloser interface that will handle the key exchange and the
// fragmentation of data into multiple chunks to enable data streaming.
package secureconn

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/nacl/box"
)

var (
	// ErrReplayAttack is the error returned when reading from a secure
	// connection and a replay attack is detected. Callers can continue to read
	// data and the replayed packet will be safely ignored.
	ErrReplayAttack = errors.New("reply attack detected")

	// ErrVerification means that the encrypted data read failed signature
	// verification and will not be decrypted.
	ErrVerification = errors.New("verification error")

	// ErrPacketTooLarge is the error returned when the packet being read or
	// written is too large (above MaxPacketSize)
	ErrPacketTooLarge = errors.New("packet data too large")

	// ErrInvalidChunkSize is the error returned when setting a chunk size that
	// is below MinimumStreamingChunkSize or above MaxPacketSize.
	ErrInvalidChunkSize = fmt.Errorf("chunk size must be between >= %d and <= %d",
		MinimumStreamingChunkSize, MaxPacketSize)
)

const (
	// Overhead is the how many extra bytes are added to each packet sent.
	Overhead = 24 + box.Overhead

	// MaxPacketSize is the maximum packet length minus overhead.
	MaxPacketSize = 32768 - Overhead

	// DefaultStreamingChunkSize is the default maximum size of packets sent
	// minus overhead.
	DefaultStreamingChunkSize = 1024 - Overhead

	// MinimumStreamingChunkSize is the very minimum chunk size allowed, but
	// is not recommended due to the overhead that will be attached to each
	// chunk.
	MinimumStreamingChunkSize = Overhead + 1
)

// SecureConn represents a secured connection. It implements the ReadWriteCloser
// interface.
type SecureConn struct {
	io.Reader
	io.Writer
	io.Closer
	chunkSize int
}

// New wraps a ReadWriterCloser and uses the underlying reader and writer to
// perform a key exchange and then returns a secured connection.
func New(conn io.ReadWriteCloser) (*SecureConn, error) {
	// Generate key pair
	priv, pub, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	// Send our public key
	// This is done in a goroutine so that read/writes can happen concurrently
	// in the event that the reader and writer are directly connected (for
	// example via io.Pipe)
	done := make(chan error)
	go func() {
		_, err := conn.Write(pub[:])
		done <- err
	}()

	// Read their public key
	key := &[32]byte{}
	if _, err := io.ReadFull(conn, key[:]); err != nil {
		return nil, err
	}
	if err := <-done; err != nil {
		return nil, err
	}

	// Return a secured connection
	return &SecureConn{
		NewSecureReader(conn, priv, key),
		NewSecureWriter(conn, priv, key),
		conn,
		DefaultStreamingChunkSize,
	}, nil
}

// Write implements the standard Write interface: it writes data as NaCl
// encrypted boxed messages to the underlying writer. Boxed messages have to be
// fully received to be decrypted, so data is fragmented into multiple chunks,
// each chunk being an individually boxed message. This gives the secure
// connection the ability to stream data.
//
// The maximum size of each chunk can be controlled with SetStreamingChunkSize.
func (sc *SecureConn) Write(p []byte) (n int, err error) {
	if sc.chunkSize == 0 {
		return sc.Writer.Write(p)
	}

	var written int
	for {
		max := len(p)
		if max > DefaultStreamingChunkSize {
			max = DefaultStreamingChunkSize
		}

		written, err = sc.Writer.Write(p[:max])
		n += written

		if err != nil {
			return
		}

		if len(p) == max {
			break
		}
		p = p[max:]
	}

	return
}

// SetStreamingChunkSize sets the maximum size for each chunk sent when
// streaming data.
//
// A sensible chunk size should be chosen as each chunk sent will include
// overhead.
//
// Streaming, and therefore the chunking of data, can be disabled by providing a
// chunk size of 0.
func (sc *SecureConn) SetStreamingChunkSize(size int) error {
	if size != 0 && (size < MinimumStreamingChunkSize || size > MaxPacketSize) {
		return ErrInvalidChunkSize
	}

	sc.chunkSize = size
	return nil
}

// SecureReader wraps and secures an io.Reader.
type SecureReader struct {
	r   io.Reader // underlying reader
	key [32]byte  // shared key
	box []byte    // encrypted box data
	msg []byte    // decrypted box data remaining
	inc int       // counter
}

func (sr *SecureReader) readBox(p []byte) (err error) {
	var nonce [24]byte

	_, err = io.ReadFull(sr.r, nonce[:])
	if err != nil {
		return
	}

	length := int(binary.LittleEndian.Uint16(nonce[:2])) + box.Overhead
	count := int(binary.LittleEndian.Uint64(nonce[2:]))

	if length > MaxPacketSize {
		return ErrPacketTooLarge
	}

	if count <= sr.inc {
		return ErrReplayAttack
	}
	sr.inc = count

	// Use buffer provided as scratch space if large enough
	if len(p) >= length {
		sr.box = p[:length]
	} else {
		sr.box = make([]byte, length)
	}

	// Read encrypted box data
	_, err = io.ReadFull(sr.r, sr.box)
	if err != nil {
		return
	}

	// Decrypt and verify box
	var verified bool
	sr.msg, verified = box.OpenAfterPrecomputation(nil, sr.box, &nonce, &sr.key)
	if !verified {
		return ErrVerification
	}

	return
}

// Read reads data from the secured connection.
func (sr *SecureReader) Read(p []byte) (n int, err error) {
	// NaCl's box needs to be available in full before it can be decrypted and
	// there's no guarantee that len(p) is large enough. Encrypted data is read
	// in full and decrypted data is copied over to the caller's buffer in
	// subsequent reads.
	if sr.msg == nil {
		err = sr.readBox(p)
		if err != nil {
			return
		}
	}

	n = copy(p, sr.msg)
	sr.msg = sr.msg[n:]
	if len(sr.msg) == 0 {
		sr.msg = nil
	}

	return
}

// SecureWriter wraps and secures an io.Writer.
type SecureWriter struct {
	w   io.Writer // underlying writer
	key [32]byte  // shared key
	inc int       // counter
}

// Write writes data to the secure connection.
func (sw *SecureWriter) Write(p []byte) (n int, err error) {
	if len(p) > MaxPacketSize {
		return 0, ErrPacketTooLarge
	}

	// Nonce: uint16(len(p)) || uint64(inc) || 14 bytes random data
	var nonce [24]byte
	sw.inc++
	binary.LittleEndian.PutUint16(nonce[:2], uint16(len(p)))
	binary.LittleEndian.PutUint64(nonce[6:], uint64(sw.inc))
	_, err = rand.Read(nonce[10:])
	if err != nil {
		return
	}

	// Write encrypted data to underlying writer
	n, err = sw.w.Write(box.SealAfterPrecomputation(nonce[:], p, &nonce, &sw.key))

	// Return written bytes, minus encryption overhead and nonce lengths
	return n - Overhead, err
}

// NewSecureReader instantiates a new SecureReader.
func NewSecureReader(r io.Reader, priv, pub *[32]byte) *SecureReader {
	sr := &SecureReader{r: r}
	box.Precompute(&sr.key, priv, pub)

	return sr
}

// NewSecureWriter instantiates a new SecureWriter.
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) *SecureWriter {
	sw := &SecureWriter{w: w}
	box.Precompute(&sw.key, priv, pub)

	return sw
}
