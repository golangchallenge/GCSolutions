package main

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	"golang.org/x/crypto/nacl/box"
)

type key [32]byte

func (k *key) String() string {
	return hex.EncodeToString(k[:])
}

type nonce [24]byte

func (n *nonce) String() string {
	return fmt.Sprintf("%s (seq %d)", hex.EncodeToString(n[:16]), binary.LittleEndian.Uint64(n[16:]))
}

// NewSecureReader instantiates a new secure reader. A secure reader is safe to
// use from multiple goroutines.
func NewSecureReader(r io.Reader, priv, peersPub *key) io.Reader {
	sharedKey := computeSharedKey(priv, peersPub)

	v(3).Printf("reader: shared key %v", sharedKey)

	return &secureReader{
		r: r,

		sharedKey: sharedKey,
	}
}

type secureReader struct {
	r io.Reader // underlying reader

	sharedKey *key // shared key computed from priv and peer's public key

	mu  sync.Mutex // ensures that only one read happens at any time
	seq uint64     // a seq number to prevent replay attack
}

func (sr *secureReader) Read(p []byte) (int, error) {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	nonce := new(nonce)
	if _, err := io.ReadFull(sr.r, nonce[:]); err != nil {
		if err == io.EOF {
			return 0, err
		}
		return 0, fmt.Errorf("read nonce: %v", err)
	}

	v(3).Printf("reader: read nonce %v", nonce)

	if err := sr.checkAndUpdateSeq(nonce); err != nil {
		return 0, err
	}

	var size uint16
	if err := binary.Read(sr.r, binary.LittleEndian, &size); err != nil {
		return 0, fmt.Errorf("read size: %v", err)
	}

	v(3).Printf("reader: size of payload %v", size)

	buf := make([]byte, size)
	if _, err := io.ReadFull(sr.r, buf); err != nil {
		return 0, fmt.Errorf("read encrypted data: %v", err)
	}

	v(3).Printf("reader: read %v bytes", len(buf))

	opened, ok := box.OpenAfterPrecomputation(nil, buf, (*[24]byte)(nonce), (*[32]byte)(sr.sharedKey))
	if !ok {
		return 0, errors.New("could not authenticate data")
	}

	v(3).Printf("reader: opened the boxed data, len %v", len(opened))

	n := copy(p, opened)
	return n, nil
}

func (sr *secureReader) checkAndUpdateSeq(n *nonce) error {
	seq := binary.LittleEndian.Uint64(n[16:])
	if seq <= sr.seq {
		return fmt.Errorf("read seq %v, expected >= %v", seq, sr.seq)
	}
	sr.seq = seq
	return nil
}

// NewSecureWriter instantiates a new secure writer. A secure writer is safe to
// use from multiple goroutines.
func NewSecureWriter(w io.Writer, priv, peersPub *key) io.Writer {
	sharedKey := computeSharedKey(priv, peersPub)

	v(3).Printf("writer: shared key %v", sharedKey)

	return &secureWriter{
		w: w,

		sharedKey: sharedKey,
	}
}

type secureWriter struct {
	w io.Writer // underlying writer

	sharedKey *key // shared key computed from priv and peer's public key

	mu  sync.Mutex // ensures that only one write happens at any time
	seq uint64     // a seq number to prevent replay attack
}

func (sw *secureWriter) Write(p []byte) (int, error) {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	nonce, err := sw.nextNonce()
	if err != nil {
		return 0, fmt.Errorf("generate nonce: %v", err)
	}

	if _, err = sw.w.Write(nonce[:]); err != nil {
		return 0, fmt.Errorf("write nonce: %v", err)
	}

	v(3).Printf("writer: wrote nonce %v", nonce)

	sealed := box.SealAfterPrecomputation(nil, p, (*[24]byte)(nonce), (*[32]byte)(sw.sharedKey))

	v(3).Printf("writer: sealed data len %v", len(sealed))

	size := uint16(len(sealed))
	if err := binary.Write(sw.w, binary.LittleEndian, size); err != nil {
		return 0, fmt.Errorf("write size: %v", err)
	}

	v(3).Printf("writer: wrote length of sealed data")

	if _, err := sw.w.Write(sealed); err != nil {
		return 0, fmt.Errorf("write encrypted data: %v", err)
	}

	v(3).Printf("writer: wrote sealed data")

	return len(p), nil
}

func (sw *secureWriter) nextNonce() (*nonce, error) {
	nonce := new(nonce)

	if _, err := io.ReadFull(rand.Reader, nonce[:16]); err != nil {
		return nil, err
	}

	// Increment the seq no.
	sw.seq++

	binary.LittleEndian.PutUint64(nonce[16:], sw.seq)

	return nonce, nil
}

func computeSharedKey(priv, pub *key) *key {
	sharedKey := new(key)
	box.Precompute((*[32]byte)(sharedKey), (*[32]byte)(pub), (*[32]byte)(priv))
	return sharedKey
}

// exchange exchanges the public keys of the two connected parties.
func exchange(conn net.Conn, pub *key) (*key, error) {
	if _, err := conn.Write(pub[:]); err != nil {
		return nil, fmt.Errorf("writing public key: %v", err)
	}

	peersPub := new(key)
	if _, err := io.ReadFull(conn, peersPub[:]); err != nil {
		return nil, fmt.Errorf("reading peer's public key: %v", err)
	}

	return peersPub, nil
}
