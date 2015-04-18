package secure

import (
	"errors"
	"io"

	"golang.org/x/crypto/nacl/box"
)

// Reader wraps io.Reader and decrypts its content using NaCl.
type Reader struct {
	reader    io.Reader
	sharedKey [32]byte
}

// NewReader instantiates a new Reader
func NewReader(r io.Reader, priv, pub *[32]byte) *Reader {
	sr := &Reader{
		reader: r,
	}

	box.Precompute(&sr.sharedKey, pub, priv)

	return sr
}

func (sr *Reader) Read(p []byte) (n int, err error) {
	data, nonce, err := sr.readMessage()
	if err != nil {
		return 0, err
	}
	n = len(data) - box.Overhead

	buf := make([]byte, 0, n)
	buf, ok := box.OpenAfterPrecomputation(buf, data, &nonce, &sr.sharedKey)
	if !ok {
		return 0, errors.New("failed to decrypt")
	}
	copy(p[:n], buf)

	return n, nil
}

func (sr *Reader) readMessage() (data []byte, nonce [NonceLength]byte, err error) {
	data = make([]byte, 1024)
	n, err := sr.reader.Read(data)
	if err != nil {
		return nil, nonce, err
	}

	if n <= box.Overhead+NonceLength {
		return nil, nonce, errors.New("incomplete message")
	}

	data = data[:n-NonceLength]
	copy(nonce[:], data[n-NonceLength:n])

	return data, nonce, nil
}
