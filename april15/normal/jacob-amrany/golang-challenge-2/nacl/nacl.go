package nacl

import (
	"crypto/rand"
	"errors"
	"io"
	"net"

	"golang.org/x/crypto/nacl/box"
)

type NaclReader struct {
	r  io.Reader
	sh *[32]byte
}

type NaclWriter struct {
	w  io.Writer
	sh *[32]byte
}

type NaclRWC struct {
	Cl io.Closer
	R  io.Reader
	W  io.Writer
}

func (rwc NaclRWC) Read(b []byte) (int, error) {
	return rwc.R.Read(b)
}

func (rwc NaclRWC) Write(b []byte) (int, error) {
	return rwc.W.Write(b)
}

func (rwc NaclRWC) Close() error {
	if rwc.Cl != nil {
		er := rwc.Cl.Close()
		return er
	}
	return nil
}

func (nr NaclReader) Read(b []byte) (n int, err error) {
	var nonce [24]byte

	buf := make([]byte, 1028)
	n, err = nr.r.Read(buf)
	if err != nil {
		return n, err
	}
	copy(nonce[:], buf[:24])
	buf = buf[24:n]
	dec, ok := box.OpenAfterPrecomputation(nil, buf, &nonce, nr.sh)
	if !ok {
		err = errors.New("Something went wrong with decrypting")
		return n, err
	}
	if len(dec) != len(buf)-box.Overhead {
		err = errors.New("Number of bytes is not equal to the correct number")
		return n, err
	}

	m := copy(b, dec)
	return m, err
}

func (nw NaclWriter) Write(b []byte) (n int, err error) {
	var nonce [24]byte
	rand.Read(nonce[:])

	enc := box.SealAfterPrecomputation(nonce[:], b, &nonce, nw.sh)
	n, err = nw.w.Write(enc)

	if n != len(b)+box.Overhead+24 {
		return n, errors.New("Didn't encrypt the correct amount of bytes")
	}

	return n, err
}

func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	var sh [32]byte

	nr := NaclReader{
		r:  r,
		sh: &sh,
	}

	box.Precompute(&sh, pub, priv)

	return nr
}
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	var sh [32]byte

	nw := NaclWriter{
		w:  w,
		sh: &sh,
	}

	box.Precompute(&sh, pub, priv)

	return nw
}

func NewRWC(conn net.Conn, priv, peer *[32]byte) NaclRWC {
	r := NewSecureReader(conn, priv, peer)
	w := NewSecureWriter(conn, priv, peer)

	rwc := NaclRWC{
		W:  w,
		R:  r,
		Cl: conn,
	}
	return rwc
}
