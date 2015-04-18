package main

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"sync"

	"golang.org/x/crypto/nacl/box"
)

const (
	// Maximum size of plain text message.
	kMaxMessage = 32 * 1024
)

var (
	errMessageTooBig = fmt.Errorf("message too big")
	errDecription    = fmt.Errorf("decription failure")
)

func GenerateKey() (publicKey, privateKey *[32]byte, err error) {
	return box.GenerateKey(rand.Reader)
}

func MakeSharedKey(sharedKey, peersPublicKey, privateKey *[32]byte) *[32]byte {
	if sharedKey == nil {
		sharedKey = new([32]byte)
	}
	box.Precompute(sharedKey, peersPublicKey, privateKey)
	return sharedKey
}

type coder struct {
	sync.Mutex
	//counter   uint32
	sharedKey *[32]byte
	nonce     [24]byte
	buffer    []byte // raw buffer for size and encripted chunk. For reuse.
}

func (c *coder) getBuffer() []byte {
	if c.buffer == nil {
		c.buffer = make([]byte, 2+box.Overhead+kMaxMessage)
	}
	return c.buffer
}

func (c *coder) encode(msg []byte, out io.Writer) (err error) {
	if _, err = rand.Read(c.nonce[:]); err != nil {
		return
	}
	// Trim length to size only. Box will append encoded message.
	buffer := c.getBuffer()[:2]
	m := len(msg) + box.Overhead
	binary.BigEndian.PutUint16(buffer, uint16(m))
	box.SealAfterPrecomputation(buffer, msg, &c.nonce, c.sharedKey)
	if _, err = out.Write(c.nonce[:]); err != nil {
		return
	}
	_, err = out.Write(buffer[:2+m])
	return
}

func (c *coder) decode(msg []byte, in io.Reader) (n int, err error) {
	if _, err = io.ReadFull(in, c.nonce[:]); err != nil {
		return
	}
	buffer := c.getBuffer()
	if _, err = io.ReadFull(in, buffer[:2]); err != nil {
		return
	}
	m := int(binary.BigEndian.Uint16(buffer))
	if m > kMaxMessage+box.Overhead {
		return 0, errMessageTooBig
	}
	buffer = buffer[:m]
	if _, err = io.ReadFull(in, buffer); err != nil {
		return
	}
	_, ok := box.OpenAfterPrecomputation(msg, buffer, &c.nonce, c.sharedKey)
	if !ok {
		return 0, errDecription
	}
	n = m - box.Overhead
	return
}

type boxReader struct {
	coder
	r       io.Reader
	decoded []byte // buffer for decoded message. Reused.
	off     int    // offset inside of decoded buffer
}

func (r *boxReader) resetDecoded() {
	if r.decoded == nil {
		r.decoded = make([]byte, kMaxMessage)
	}
	r.decoded = r.decoded[:0]
	r.off = 0
}

func (r *boxReader) set(in io.Reader, sharedKey *[32]byte) {
	r.r = in
	r.sharedKey = sharedKey
}

func (r *boxReader) Read(b []byte) (n int, err error) {
	if len(b) == 0 {
		return
	}
	r.Lock()
	defer r.Unlock()

	if r.off < len(r.decoded) {
		nc := copy(b, r.decoded[r.off:])
		b = b[nc:]
		n += nc
		r.off += nc
		return
	}
	count := 0
	for len(b) > 0 {
		r.resetDecoded()
		var m int
		m, err = r.decode(r.decoded, r.r)
		if err != nil {
			return
		}
		count++
		r.decoded = r.decoded[:m]
		nc := copy(b, r.decoded)
		b = b[nc:]
		n += nc
		r.off += nc
		return
	}
	return
}

// Implementation of io.WriterTo interface. This allows to use io.Copy without
// buffer.
func (r *boxReader) WriteTo(w io.Writer) (n int64, err error) {
	r.Lock()
	defer r.Unlock()

	var nc int
	if r.off < len(r.decoded) {
		nc, err = w.Write(r.decoded[r.off:])
		n += int64(nc)
		r.off += nc
		if err != nil {
			return
		}
	}
	for err == nil {
		r.resetDecoded()
		var m int
		m, err = r.decode(r.decoded, r.r)
		if err != nil {
			return
		}
		r.decoded = r.decoded[:m]
		nc, err = w.Write(r.decoded)
		n += int64(nc)
		r.off += nc
		return
	}
	return
}

var _ io.WriterTo = &boxReader{}

type boxWriter struct {
	coder
	w io.Writer
}

func (w *boxWriter) set(out io.Writer, sharedKey *[32]byte) {
	w.w = out
	w.sharedKey = sharedKey
}

func (w *boxWriter) Write(b []byte) (n int, err error) {
	if len(b) == 0 {
		return
	}
	w.Lock()
	defer w.Unlock()

	m := len(b)
	for m > 0 {
		if m > kMaxMessage {
			m = kMaxMessage
		}
		if err = w.encode(b[:m], w.w); err != nil {
			return
		}
		n += m
		b = b[m:]
		m = len(b)
	}
	return
}

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	return &boxReader{
		coder: coder{sharedKey: MakeSharedKey(nil, pub, priv)},
		r:     r,
	}
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	return &boxWriter{
		coder: coder{sharedKey: MakeSharedKey(nil, pub, priv)},
		w:     w,
	}
}
