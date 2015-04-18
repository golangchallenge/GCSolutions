package main

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"

	"golang.org/x/crypto/nacl/box"
)

var (
	// MaxMessageLength is the maximum size of a message. This is to prevent memory allocation attacks.
	// In this case, we use 32kb - 1 since that's the challeges max length.
	MaxMessageLength  = 31999
	nonceHeaderLength = 24
)

// CryptoRandomReader generates crypto random data
type CryptoRandomReader struct{}

// Read will put random data into p, it will try to fill p entirely with random data
func (r *CryptoRandomReader) Read(p []byte) (n int, err error) {
	return rand.Read(p)
}

// Message is a representation of an indivudal message that can be encoded and decoded
type Message struct {
	// Data is the underlying data
	Data []byte
}

// Encoder encrypts a Message and sends it over a Writer
type Encoder struct {
	w         io.Writer
	sharedKey *[32]byte
}

// NewEncoder allocates an Encoder and initializes it for you.
func NewEncoder(w io.Writer, sharedKey *[32]byte) *Encoder {
	enc := &Encoder{}
	enc.w = w
	enc.sharedKey = sharedKey

	return enc
}

// Encode encrypts a Message and sends it over a Writer
func (enc *Encoder) Encode(msg *Message) error {
	// rand.Read is guaranteed to read 24 bytes because it calls ReadFull under the covers
	nonceBytes := make([]byte, 24)
	_, err := rand.Read(nonceBytes)
	if err != nil {
		return err
	}

	// We create a fixed array to copy the nonceBytes (there is no way to convert from slice to fixed array without copying)
	var nonce [24]byte
	copy(nonce[:], nonceBytes[:])

	// box.SealAfterPrecomputation appends the encrypted data to it out and returns it
	// We pass nonceBytes to the out parameter so we get returned data in the form [nonce][encryptedData]
	data := box.SealAfterPrecomputation(nonceBytes, msg.Data, &nonce, enc.sharedKey)

	// Prepend the length to our data so the reader knows how much room to make when reading
	var length = uint32(len(data))
	err = binary.Write(enc.w, binary.BigEndian, length)
	if err != nil {
		return nil
	}

	_, err = enc.w.Write(data)
	if err != nil {
		return err
	}

	return nil
}

// Decoder decrypts data from a Reader. The data is expected to be encoded by Encoder
type Decoder struct {
	r         io.Reader
	sharedKey *[32]byte
}

// NewDecoder allocates an Encoder and initializes it for you.
func NewDecoder(r io.Reader, sharedKey *[32]byte) *Decoder {
	dec := &Decoder{}
	dec.r = r
	dec.sharedKey = sharedKey

	return dec
}

// Decode decrypts a Message from the underlying Reader and stores it in m
func (dec *Decoder) Decode(m *Message) error {
	// Length is the length of the encrypted data (including box.Overhead)
	var length uint32
	err := binary.Read(dec.r, binary.BigEndian, &length)
	if err != nil {
		return err
	}
	if length <= 0 {
		return fmt.Errorf("invalid length (len:%d) for encrypted data", length)
	}
	// restrict length to stop memory allocation attack
	maxLength := uint32(MaxMessageLength + nonceHeaderLength + box.Overhead)
	if length > maxLength {
		return fmt.Errorf("length of encrypted data is too large (len:%d max: %d)", length, maxLength)
	}

	// To be able to decrypt properly, we must receive all the data that we encrypted with
	data := make([]byte, length)
	_, err = io.ReadFull(dec.r, data)
	if err != nil {
		return err
	}

	var nonce [24]byte
	copy(nonce[:], data[0:24])

	// OpenAfterPrecomputation appends to out and returns the appended data
	data, ok := box.OpenAfterPrecomputation(nil, data[24:], &nonce, dec.sharedKey)

	// If ok is false, we have failed to decrypt properly
	// Usually this is because the encrypted data is malformed
	if !ok {
		return fmt.Errorf("failed to decrypt box! Encrypted data is likely malformed")
	}

	m.Data = data

	return nil
}

// SecureReadWriteCloser implements a secure ReadWriteCloser using public-key cryptography
type SecureReadWriteCloser struct {
	sr  *SecureReader
	sw  *SecureWriter
	rwc io.ReadWriteCloser
}

// Init initializes a SecureWriteCloser with a private and public key
// rwc is an underlying ReadWriteCloser we want to make secure
// priv is your private key
// pub is the public key of the party you're trying to communicate with
func (srwc *SecureReadWriteCloser) Init(rwc io.ReadWriteCloser, priv, pub *[32]byte) {
	srwc.sr = NewSecureReader(rwc, priv, pub)
	srwc.sw = NewSecureWriter(rwc, priv, pub)
	srwc.rwc = rwc
}

// Read decrypts from the underlying stream and writes it to p []byte
// p is expected to be big enough to hold the entire decrypted message, if it's not,
// Read writes as much as it can and discards the rest of the message.
func (srwc *SecureReadWriteCloser) Read(msg []byte) (n int, err error) {
	return srwc.sr.Read(msg)
}

// ReadMsg decrypts an entire box from the underlying stream and returns it
func (srwc *SecureReadWriteCloser) ReadMsg() (msg *Message, err error) {
	return srwc.sr.ReadMsg()
}

// Write encrypts p []byte and sends it to the underlying stream
func (srwc *SecureReadWriteCloser) Write(msg []byte) (n int, err error) {
	return srwc.sw.Write(msg)
}

// Close closes the underlying stream
func (srwc *SecureReadWriteCloser) Close() error {
	return srwc.rwc.Close()
}

// NewSecureReadWriteCloser allocates a SecureReadWriteCloser for you and initializes it
func NewSecureReadWriteCloser(r io.ReadWriteCloser, priv, pub *[32]byte) *SecureReadWriteCloser {
	srwc := &SecureReadWriteCloser{}
	srwc.Init(r, priv, pub)
	return srwc
}

// SecureReader decrypts from a stream securely using public-key cryptography
type SecureReader struct {
	dec *Decoder
}

// NewSecureReader is a convenient helper method that allocates and initializes a secure reader for you
// r is the underlying stream to read securely from
// priv is your private key
// pub is the public key of who you're communicating with
func NewSecureReader(r io.Reader, priv, pub *[32]byte) *SecureReader {
	sr := &SecureReader{}
	sr.Init(r, priv, pub)
	return sr
}

// Init initializes our Reader
// r is the underlying stream to read securely from
// priv is your private key
// pub is the public key of who you're communicating with
func (sr *SecureReader) Init(r io.Reader, priv, pub *[32]byte) {
	var sharedKey [32]byte
	box.Precompute(&sharedKey, pub, priv)
	sr.dec = NewDecoder(r, &sharedKey)
}

// ReadMsg decrypts an entire message from the underlying stream and returns it
// ReadMsg is more effecient than calling .Read() because you don't need to preallocate
// the max message size beforehand.
func (sr *SecureReader) ReadMsg() (msg *Message, err error) {
	msg = new(Message)

	err = sr.dec.Decode(msg)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

// Read decrypts a box from the underlying stream and writes it to p []byte
// p is expected to be big enough to hold the entire decrypted message, if it's not,
// Read writes as much as it can to p []byte and discards the rest of the message.
func (sr *SecureReader) Read(p []byte) (n int, err error) {
	var msg Message
	err = sr.dec.Decode(&msg)
	if err != nil {
		return 0, err
	}

	n = copy(p, msg.Data)
	return n, nil
}

// SecureWriter encrypts data securely to a stream
type SecureWriter struct {
	enc *Encoder
}

// NewSecureWriter is a convenient helper method that allocates and initializes a secure writer for you
// w is the underlying stream to write securely to
// priv is your private key
// pub is the public key of who you're communicating with
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) *SecureWriter {
	sw := &SecureWriter{}
	sw.Init(w, priv, pub)
	return sw
}

// Init initializes our Writer.
// w is the underlying stream to write securely to
// priv is your private key
// pub is the public key of who you're communicating with
func (sw *SecureWriter) Init(w io.Writer, priv, pub *[32]byte) {
	var sharedKey [32]byte
	box.Precompute(&sharedKey, pub, priv)
	sw.enc = NewEncoder(w, &sharedKey)
}

// Write encrypts p []byte to the underlying stream.
func (sw *SecureWriter) Write(p []byte) (n int, err error) {
	err = sw.enc.Encode(&Message{Data: p})
	if err != nil {
		return 0, err
	}

	// If encoding is successful, we're guaranteed that all the data was written
	return len(p), nil
}
