package main

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"net"
	"sync"

	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/nacl/box"
)

var (
	// ErrMsgTooLong is returned by SecureWriter.Write when the passed message
	// is too long for a single packet
	ErrMsgTooLong = errors.New("the message can't be longer than (2^32 - 1) bytes")

	// ErrSeqWrapped is returned by SecureWriter.Write when the sequence number
	// is almost at the maximum value for a uint32. This is a protection against
	// counter wrapping attacks.
	ErrSeqWrapped = errors.New("the sequence number overflowed its limit")

	// ErrFailedVerify is returned by SecureReader.Read when NaCl verification fails
	ErrFailedVerify = errors.New("received data failed verification")

	// ErrOutOfOrder is returned by SecureReader.Read when it receives a
	// packet with an unexpected sequence number
	ErrOutOfOrder = errors.New("received data was not sequential")
)

// PacketHdr is the data sent by SecureWriter before each encrypted payload
type PacketHdr struct {
	Nonce  [24]byte
	Length uint32
}

// SecureWriter encrypts data with NaCl and writes it to the underlying Writer.
type SecureWriter struct {
	mu  sync.Mutex
	wr  io.Writer
	seq uint32
	key *[32]byte
}

// NewSecureWriter instantiates a new SecureWriter. w is the target Writer, priv
// is the local private key for authentication and pub is the peer's public key
// for encryption.
//
// IMPORTANT: never reuse the same priv in another SecureWriter. Even if the
// nonce is random, it opens to replay attacks since the attacker can replay all
// the packets of a session using the same (priv, pub). It is safe and expected
// to reuse it in a corresponding SecureReader.
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	sw := &SecureWriter{
		wr:  w,
		key: new([32]byte),
	}
	box.Precompute(sw.key, pub, priv)

	// Compute our public key and compare it to the other party's to decide who
	// should use odd sequence numbers (has higher pubkey -> write odd seq)
	ourPub := new([32]byte)
	curve25519.ScalarBaseMult(ourPub, priv)
	if bytes.Compare(ourPub[:], pub[:]) > 0 {
		sw.seq = 1
	}

	return sw
}

// Write encrypts and writes data to the underlying Writer.
//
// Note: SecureWriter does not perform any buffering, and it incurs in a small
// overhead for each call to Write. You might want to wrap it in a bufio.Writer.
func (w *SecureWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	hdr := PacketHdr{}

	// 1. Generate a random nonce
	if _, err := rand.Read(hdr.Nonce[:]); err != nil {
		return 0, err
	}

	// 2. Prepare the data to be encrypted:
	// - a sequence number to avoid reordering attacks
	// - the passed plaintext
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, w.seq); err != nil {
		return 0, err
	}
	if _, err := buf.Write(p); err != nil {
		return 0, err
	}

	// 3. Calculate the total length of the encrypted payload
	if buf.Len()+box.Overhead > math.MaxUint32 {
		return 0, ErrMsgTooLong
	}
	hdr.Length = uint32(buf.Len() + box.Overhead)

	// 4. Encrypt the payload
	encData := box.SealAfterPrecomputation(nil, buf.Bytes(), &hdr.Nonce, w.key)

	// 5. Write header and payload
	if err := binary.Write(w.wr, binary.BigEndian, hdr); err != nil {
		return 0, err
	}
	if _, err := w.wr.Write(encData); err != nil {
		return 0, err
	}

	if w.seq += 2; w.seq > (math.MaxUint32 - 2) {
		return len(p), ErrSeqWrapped
	}

	return len(p), nil
}

// SecureReader reads encrypted messages from a Reader, decrypts and buffers them.
type SecureReader struct {
	mu  sync.Mutex
	rd  io.Reader
	buf []byte
	seq uint32
	key *[32]byte
	err error
}

// NewSecureReader instantiates a new SecureReader. r is the source Reader, priv
// is the local private key for decryption and pub is the peer's public key
// for authentication.
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	sr := &SecureReader{
		rd:  r,
		key: new([32]byte),
	}
	box.Precompute(sr.key, pub, priv)

	// Compute our public key and compare it to the other party's to decide who
	// should use odd sequence numbers (has higher pubkey -> write odd seq)
	ourPub := new([32]byte)
	curve25519.ScalarBaseMult(ourPub, priv)
	if bytes.Compare(pub[:], ourPub[:]) > 0 {
		sr.seq = 1
	}

	return sr
}

// Read decrypts data into p. It reads at most one message from the network,
// hence n may be less than len(p) like for bufio.Reader. Use io.ReadFull to
// fill a buffer.
func (r *SecureReader) Read(p []byte) (n int, err error) {
	// Acquire a lock to avoid multiple Read() interleaving calls to the
	// underlying r.rd.Read() and out-of-order reads
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.err != nil {
		return 0, r.err
	}

	if len(r.buf) < len(p) {
		r.err = r.readPacket()
	}

	n = copy(p, r.buf)
	r.buf = r.buf[n:]
	return n, r.err
}

// readPacket reads a packet from r.rd into r.buf. EOF is returned if
// r.rd.Read() returns it. Nothing is appended to r.buf if err != nil.
func (r *SecureReader) readPacket() error {
	// 1. Read the packet header
	hdr := &PacketHdr{}
	if err := binary.Read(r.rd, binary.BigEndian, hdr); err != nil {
		return err
	}

	// 2. Read the encrypted data
	encData := make([]byte, hdr.Length)
	if _, err := io.ReadFull(r.rd, encData); err != nil {
		return err
	}

	// 3. Decrypt and verify the data
	b, ok := box.OpenAfterPrecomputation(nil, encData, &hdr.Nonce, r.key)
	if !ok {
		return ErrFailedVerify
	}
	buf := bytes.NewReader(b)

	// 4. Extract the sequence number and match it to the internal counter
	var seq uint32
	if err := binary.Read(buf, binary.BigEndian, &seq); err != nil {
		return err
	}
	if seq != r.seq {
		return ErrOutOfOrder
	}
	r.seq += 2

	// Success! Note: this append will cause the reallocation of r.buf when
	// it grows too much, and with that the old returned data will be discarded
	r.buf = append(r.buf, b[len(b)-buf.Len():]...)
	return nil
}

// SecureConn is a net.Conn wrapper that will exchange freshly generated keys
// with the SecureConn on the other side (performing no kind of checking, so
// it's vulnerable to active MitM attacks) and then initialize a SecureReader
// and a SecureWriter.
//
// It implements io.ReadWriteCloser.
type SecureConn struct {
	co   net.Conn
	priv *[32]byte
	pub  *[32]byte
	rd   io.Reader
	wr   io.Writer
}

// NewSecureConn performs the key exchange over the passed net.Conn and then
// returns a SecureConn wrapping it.
func NewSecureConn(co net.Conn) (*SecureConn, error) {
	c := &SecureConn{co: co}

	// Generate a local keypair
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	c.priv = priv

	// Send the public key to the other party
	if _, err := c.co.Write(pub[:]); err != nil {
		return nil, err
	}

	// Receive the other party's public key
	var key [32]byte
	if _, err := io.ReadFull(c.co, key[:]); err != nil {
		return nil, err
	}
	c.pub = &key

	// Initialize the SecureReader and SecureWriter
	c.wr = NewSecureWriter(c.co, c.priv, c.pub)
	c.rd = NewSecureReader(c.co, c.priv, c.pub)

	return c, nil
}

// Read exposes the SecureReader's Read method. Same caveats apply.
func (c *SecureConn) Read(p []byte) (n int, err error) {
	return c.rd.Read(p)
}

// Write exposes the SecureWriter's Write method. Same caveats apply.
func (c *SecureConn) Write(p []byte) (n int, err error) {
	return c.wr.Write(p)
}

// Close directly closes the underlying connection.
func (c *SecureConn) Close() error {
	return c.co.Close()
}
