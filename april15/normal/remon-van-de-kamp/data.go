package main

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"golang.org/x/crypto/nacl/box"
	"io"
	"io/ioutil"
	"strconv"
	"sync"
	"time"
)

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	return &SecureReader{
		reader: r,
		priv:   priv,
		pub:    pub,
	}
}

// SecureReader implements io.Reader and can be used to wrap
// another io.Reader and read encrypted messages from
// said io.Reader
type SecureReader struct {
	reader io.Reader
	priv   *[32]byte
	pub    *[32]byte
	mutex  sync.Mutex
}

// Read reads messages from the wrapped io.Reader and decodes them
func (s *SecureReader) Read(p []byte) (n int, err error) {
	// make sure there is only one goroutine reading
	// from the datastream at a time
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var nonce [24]byte
	binary.Read(s.reader, binary.LittleEndian, &nonce)

	var length int16
	binary.Read(s.reader, binary.LittleEndian, &length)

	reader := io.LimitReader(s.reader, int64(length))
	data, err := ioutil.ReadAll(reader)

	result, success := box.Open([]byte{}, data, &nonce, s.pub, s.priv)
	if !success {
		return 0, errors.New("Unable to decode incoming message")
	}

	copy(p, result[:])
	n = len(result[:])
	return
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	return &SecureWriter{
		writer: w,
		priv:   priv,
		pub:    pub,
	}
}

// SecureWriter implements io.Writer and can be used to wrap
// another io.Writer and write encrypted messages to
// said io.Writer
type SecureWriter struct {
	writer io.Writer
	priv   *[32]byte
	pub    *[32]byte
	mutex  sync.Mutex
}

// Write writes encrypted messaged to the wrapped io.Writer
func (s *SecureWriter) Write(p []byte) (n int, err error) {
	nonce := generateNonce()
	res := box.Seal([]byte{}, p, nonce, s.pub, s.priv)

	headerLength := 25 // 1 for length int16, 24 for nonce [24]byte

	n = headerLength + len(res)
	if n > 1<<15 {
		return 0, errors.New("Message too long. Please send at most 32kB")
	}

	// make sure there is only one goroutine wrting
	// to the datastream at a time
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if err = binary.Write(s.writer, binary.LittleEndian, nonce); err != nil {
		return
	}

	length := int16(len(res))
	if err = binary.Write(s.writer, binary.LittleEndian, length); err != nil {
		return
	}

	m, err := s.writer.Write(res)
	if err != nil {
		return
	}
	if m != len(res) {
		return headerLength + m, errors.New("Unable to write complete message")
	}

	return
}

// generateNonce generates a random string to be used
// as nonce for nacl/box Seal/Open
func generateNonce() *[24]byte {
	seed := strconv.Itoa(time.Now().Nanosecond())
	hashedSeed := sha256.Sum224([]byte(seed))
	var nonce [24]byte
	copy(nonce[:], hashedSeed[:])
	return &nonce
}

// SecureConnection wraps a net.Conn with an
// io.Reader that can read from that connection
// and an io.Writer that can write to the connection
type SecureConnection struct {
	reader io.Reader
	writer io.Writer
	conn   io.ReadWriteCloser
}

// NewSecureConnection instantiates a SecureConnection
// to read/write messages from
func NewSecureConnection(conn io.ReadWriteCloser, priv, pub *[32]byte) *SecureConnection {
	return &SecureConnection{
		reader: NewSecureReader(conn, priv, pub),
		writer: NewSecureWriter(conn, priv, pub),
		conn:   conn,
	}
}

// Read reads from the wrapped io.Reader
func (s *SecureConnection) Read(p []byte) (n int, err error) {
	return s.reader.Read(p)
}

// Write writes to the wrapped io.Writer
func (s *SecureConnection) Write(p []byte) (n int, err error) {
	return s.writer.Write(p)
}

// Close closes the wrapped net.Conn
func (s *SecureConnection) Close() error {
	return s.conn.Close()
}
