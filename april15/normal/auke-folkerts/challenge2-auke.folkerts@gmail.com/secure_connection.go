package main

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"golang.org/x/crypto/nacl/box"
	"io"
)

const (
	nonceLen = 24
	keyLen   = 32
)

// NewSecureReader instantiates a new SecureReader.
func NewSecureReader(r io.Reader, priv, pub *[keyLen]byte) io.Reader {
	return &secureReader{r: r, pub: pub, priv: priv, buf: nil}
}

type secureReader struct {
	r         io.Reader
	pub, priv *[keyLen]byte
	buf       []byte
}

// Read reads data from a secured connection. It requires the SecureReader
// to be initialized via either NewSecureReader or NewSecureConnection
func (sr *secureReader) Read(p []byte) (int, error) {
	var size int16
	var nonce [nonceLen]byte

	// If the calling application requests less bytes than a full
	// decrypted message, we need to store the remaining bytes.

	// If there are remaining bytes, return those first.
	if sr.buf != nil {
		n := copy(p, sr.buf)

		if n >= len(sr.buf) {
			sr.buf = nil
		} else {
			sr.buf = sr.buf[n:]
		}
		return n, nil
	}

	// We had no remaining buffer, receive a new encrypted packet
	if _, err := io.ReadFull(sr.r, nonce[:]); err != nil {
		return 0, err
	}
	if err := binary.Read(sr.r, binary.BigEndian, &size); err != nil {
		return 0, err
	}
	recvbox := make([]byte, size)
	if _, err := io.ReadFull(sr.r, recvbox); err != nil {
		return 0, err
	}
	// Decrypt the packet
	buf, ok := box.Open(nil, recvbox, &nonce, sr.pub, sr.priv)

	if !ok {
		return 0, errors.New("Failed to decrypt message")
	}

	// If the caller requests enough bytes to recv the whole packet...
	if len(p) >= len(buf) {
		return copy(p, buf), nil
	}

	// ... else return at most len(p) bytes and store the rest.
	n := copy(p, buf)
	sr.buf = make([]byte, len(buf)-n)
	copy(sr.buf, buf[n:])
	return n, nil
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[keyLen]byte) io.Writer {
	return &secureWriter{w: w, pub: pub, priv: priv}
}

type secureWriter struct {
	w         io.Writer
	pub, priv *[keyLen]byte
}

// Write writes encryped data. It generates a random nonce and sends this
// along with the encrypted data so the receiving side can decrypt it.
//   It uses bytes.Buffer to pool the various elements into one write. This
// improves performance (and avoids the race in TestSecureDial where the listener
// closes the connection after reading only the first segment (ie. the nonce))
//  Write requires the SecureWriter to be initialized through either
// NewSecureWriter or NewSecureConnection.
func (sw *secureWriter) Write(p []byte) (int, error) {
	var out []byte
	var nonce [nonceLen]byte
	var buf bytes.Buffer

	if _, err := rand.Read(nonce[:]); err != nil {
		return 0, err
	}
	buf.Write(nonce[:]) // bytes.Buffer.Write never returns error

	// Encrypt the data with the generated nonce.
	out = box.Seal(nil, p, &nonce, sw.pub, sw.priv)

	// send encrypted size. By definition, messages never exceed 32768
	// bytes so an in16 suffices.
	if err := binary.Write(&buf, binary.BigEndian, int16(len(out))); err != nil {
		return 0, err
	}

	buf.Write(out) // bytes.Buffer.Write never returns error
	n, err := buf.WriteTo(sw.w)
	return int(n), err // the return value n of WriteTo always fits in an int
}

// NewSecureConnection initializes a secure connection.
func NewSecureConnection(conn io.ReadWriteCloser) (io.ReadWriteCloser, error) {

	// Generate a new Public/Private keypair
	mypub, mypriv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return &secureConnection{}, err
	}
	var otherpub [keyLen]byte

	// Perform key exchange: Send key (asynchronously)
	retval := make(chan error)
	go func() {
		if _, err := conn.Write(mypub[:]); err != nil {
			retval <- err
		} else {
			retval <- nil
		}
	}()

	// Perform key exchange: Read key (blocking)
	if _, err := io.ReadFull(conn, otherpub[:]); err != nil {
		return &secureConnection{}, err
	}

	// Cascade any error from sending the key
	if <-retval != nil {
		return &secureConnection{}, err
	}

	// Initialize the newly secured connection.
	secureConn := &secureConnection{
		conn:         conn,
		secureReader: secureReader{r: conn, priv: mypriv, pub: &otherpub},
		secureWriter: secureWriter{w: conn, priv: mypriv, pub: &otherpub}}
	return secureConn, nil
}

// A SecureConnection consists of a reader and writer interface.  Since
// the connection must be closable, also store the actual connection.
type secureConnection struct {
	conn io.Closer
	secureReader
	secureWriter
}

// Close gracefully closes the SecureConnection.
func (s *secureConnection) Close() error {
	return s.conn.Close()
}
