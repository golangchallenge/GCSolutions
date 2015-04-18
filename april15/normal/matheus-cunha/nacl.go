package main

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"golang.org/x/crypto/nacl/box"
	"io"
	"log"
)

var clientID int32

type SecureReader struct {
	r     io.Reader
	priv  *[32]byte
	pub   *[32]byte
	nonce int32
}

type SecureWriter struct {
	w     io.Writer
	priv  *[32]byte
	pub   *[32]byte
	nonce int32
}

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	return &SecureReader{r, priv, pub, clientID}
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	sw := &SecureWriter{w, priv, pub, clientID}
	clientID++
	return sw
}

func GenerateKey() (publicKey, privateKey *[32]byte, err error) {
	return box.GenerateKey(rand.Reader)
}

func (sr *SecureReader) Read(p []byte) (n int, err error) {
	n, err = sr.r.Read(p)
	if err == nil {
		var out []byte
		//log.Println("before Open ", p[0:n])
		decrypted, b := box.Open(out, p[0:n], sr.getNonce(), sr.pub, sr.priv)
		if b {
			copy(p, decrypted)
			return len(decrypted), nil
		}
	}
	return
}

func (sw *SecureWriter) Write(p []byte) (n int, err error) {
	var out []byte
	//log.Println("before Seal ", p)
	encrypted := box.Seal(out, p, sw.getNonce(), sw.pub, sw.priv)
	//log.Println("after Seal ", encrypted, out)
	return sw.w.Write(encrypted)
}

func (sr *SecureReader) Close() error {
	return sr.Close()
}

func (sw *SecureWriter) Close() error {
	return sw.Close()
}

func writeInt32(i int32) (nonce *[24]byte) {
	//will store nonce byte array
	buf := new(bytes.Buffer)

	//retuning the [24]byte representation of sw.nonce
	err := binary.Write(buf, binary.LittleEndian, i)
	if err != nil {
		log.Println("binary.Write failed:", err)
		return nil
	}

	out := new([24]byte)

	for i, v := range buf.Bytes()[0:24] {
		out[i] = v
	}

	return out
}

func (sw *SecureWriter) getNonce() (nonce *[24]byte) {
	//Unique nonce, this poor approach is because of main_test.go
	sw.nonce++
	return writeInt32(sw.nonce)
}

func (sr *SecureReader) getNonce() (nonce *[24]byte) {
	//Unique nonce, this poor approach is because of main_test.go
	sr.nonce++
	return writeInt32(sr.nonce)
}
