package main

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"encoding/json"
	"errors"
	"golang.org/x/crypto/nacl/box"
	"io"
	"net"
	"time"
)

// Message is a structure to hold a given message along with the matching nonce.
type Message struct {
	N [24]byte
	M []byte
}

// SecureConnection is a struct to allow readwriteclose to a connection.
type SecureConnection struct {
	r    io.Reader
	w    io.Writer
	conn net.Conn
}

// NewSecureConnection returns a SecureConnection already connected to a peer.
func NewSecureConnection(conn net.Conn) (io.ReadWriteCloser, error) {
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	conn.Write(pub[:])

	key := make([]byte, 32)
	n, err := conn.Read(key)
	if n != 32 || err != nil {
		conn.Write([]byte("Unable to make connection"))
		return nil, errors.New("Unable to make secure connection")
	}

	pubKey := &[32]byte{}
	for i, v := range key {
		pubKey[i] = v
	}

	c := &SecureConnection{}
	c.r = NewSecureReader(conn, priv, pubKey)
	c.w = NewSecureWriter(conn, priv, pubKey)
	c.conn = conn

	return c, nil
}

func (s *SecureConnection) Read(b []byte) (n int, err error) {
	return s.r.Read(b)
}

func (s *SecureConnection) Write(b []byte) (n int, err error) {
	return s.w.Write(b)
}

// Close terminates a connection.
func (s *SecureConnection) Close() error {
	return s.conn.Close()
}

// SecureWriter is an io.Writer for encrypted communications.
type SecureWriter struct {
	priv *[32]byte
	pub  *[32]byte
	w    io.Writer
}

// NewSecureWriter instantiates a new SecureWriter.
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	return &SecureWriter{priv: priv, pub: pub, w: w}
}

func (s *SecureWriter) Write(p []byte) (n int, err error) {
	var b bytes.Buffer
	g := gzip.NewWriter(&b)
	enc := json.NewEncoder(g)
	nonce := [24]byte{}
	no, _ := time.Now().GobEncode()
	for i, v := range no {
		nonce[i] = v
	}

	ans := box.Seal(nil, p, &nonce, s.pub, s.priv)
	data := &Message{N: nonce, M: ans}

	err = enc.Encode(&data)
	err = g.Close()
	n, err = s.w.Write(b.Bytes())
	return n, err
}

// SecureReader is an io.Reader for encrypted communications.
type SecureReader struct {
	priv *[32]byte
	pub  *[32]byte
	r    io.Reader
}

// NewSecureReader instantiates a new SecureReader.
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	return &SecureReader{priv: priv, pub: pub, r: r}
}

func (s *SecureReader) Read(p []byte) (n int, err error) {
	var data Message
	g, err := gzip.NewReader(s.r)
	if err != nil {
		return 0, err
	}

	err = json.NewDecoder(g).Decode(&data)
	g.Close()
	if err != nil {
		return 0, err
	}

	ans, _ := box.Open(nil, data.M, &data.N, s.pub, s.priv)

	for i, v := range ans {
		p[i] = v
	}

	return len(ans), err
}
