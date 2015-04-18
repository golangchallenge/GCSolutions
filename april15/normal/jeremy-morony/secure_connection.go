package main

import (
	"crypto/rand"
	"fmt"
	"io"
	"net"

	"golang.org/x/crypto/nacl/box"
)

type (
	naclReader struct {
		encrypted io.Reader
		priv      *[32]byte
		pub       *[32]byte
	}

	naclWriter struct {
		encrypted io.Writer
		priv      *[32]byte
		pub       *[32]byte
	}

	naclConn struct {
		io.Reader
		io.Writer
		io.Closer
	}
)

func exchangeKeys(c net.Conn, pub *[32]byte) (*[32]byte, error) {
	if _, err := c.Write(pub[:]); err != nil {
		return nil, err
	}

	theirs := &[32]byte{}
	if _, err := c.Read(theirs[:]); err != nil && err != io.EOF {
		return nil, err
	}

	return theirs, nil
}

func establishSecureConn(c net.Conn) (io.ReadWriteCloser, error) {
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	peersPub, err := exchangeKeys(c, pub)
	if err != nil {
		return nil, err
	}

	secure := &naclConn{
		NewSecureReader(c, priv, peersPub),
		NewSecureWriter(c, priv, peersPub),
		c,
	}

	return secure, nil
}

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	return naclReader{r, priv, pub}
}

func (r naclReader) Read(p []byte) (int, error) {
	buf := make([]byte, 2048)
	total, err := r.encrypted.Read(buf)
	if err != nil {
		return total, err
	}

	msg := buf[24:total]
	n := &[24]byte{}
	copy(n[:], buf[0:])

	decrypted, ok := box.Open(p[:0], msg, n, r.pub, r.priv)
	if !ok {
		return len(decrypted), fmt.Errorf("decrypt failed")
	}

	return len(decrypted), nil
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	return naclWriter{w, priv, pub}
}

func (w naclWriter) Write(p []byte) (int, error) {
	n := &[24]byte{}
	if _, err := rand.Read(n[:]); err != nil {
		return 0, err
	}

	var msg []byte
	msg = append(msg, n[:]...)
	msg = append(msg, box.Seal(nil, p, n, w.pub, w.priv)...)

	return w.encrypted.Write(msg)
}
