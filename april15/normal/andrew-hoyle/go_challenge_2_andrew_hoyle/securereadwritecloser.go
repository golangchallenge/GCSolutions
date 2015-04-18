package main

import "io"

type SecureReadWriteCloser struct {
	r io.Reader
	w io.Writer
	c io.Closer
}

func NewSecureReadWriteCloser(c io.ReadWriteCloser, myPriv, peersPub *[32]byte) io.ReadWriteCloser {
	return &SecureReadWriteCloser{
		r: NewSecureReader(c, myPriv, peersPub),
		w: NewSecureWriter(c, myPriv, peersPub),
		c: c,
	}
}

func (s *SecureReadWriteCloser) Close() error {
	return s.c.Close()
}

func (s *SecureReadWriteCloser) Read(p []byte) (int, error) {
	return s.r.Read(p)
}

func (s *SecureReadWriteCloser) Write(p []byte) (int, error) {
	return s.w.Write(p)
}
