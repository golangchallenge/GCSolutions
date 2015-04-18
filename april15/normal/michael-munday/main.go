package main

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"golang.org/x/crypto/nacl/box"
)

// maxLength is the maximum length of a message. The length of an encrypted
// message must be able to fit within an int on both 32- and 64-bit platforms.
const maxLength = (1 << 15) - 1 - box.Overhead

// A header encodes message length and nonce information. A header is sent
// unencrypted at the start of a message. The length is limited to 16-bits to
// avoid excessive memory usage when encrypting/decrypting messages.
type header struct {
	Length int16
	Nonce  [24]byte
}

// SecureReader provides decryption and implements the standard Reader
// interface.
//
// Calls to Read on a SecureReader decrypt messages written by a SecureWriter
// initialised using the corresponding key pair. SecureReader may buffer
// decrypted messages.
type SecureReader struct {
	src io.Reader
	key *[32]byte
	buf *bytes.Reader
}

// SecureWriter provides encryption and implements the standard Writer
// interface.
//
// Calls to Write on a SecureWriter encrypt messages which can be decrypted by a
// SecureReader initialised using the corresponding key pair.
type SecureWriter struct {
	dst io.Writer
	key *[32]byte
}

// min returns the lowest value given.
func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

// newHeader creates a header for a message with the given length (in bytes).
// The nonce will be generated from rand.Reader.
func newHeader(length int) (*header, error) {
	if length > maxLength {
		err := fmt.Errorf(
			"message size (%d bytes) is greater than %d bytes",
			length,
			maxLength,
		)
		return nil, err
	}
	h := new(header)
	h.Length = int16(length) + box.Overhead
	if _, err := io.ReadFull(rand.Reader, h.Nonce[:]); err != nil {
		return nil, err
	}
	return h, nil
}

// WriteTo serialises the header and writes it out to the given Writer. WriteTo
// implements the standard WriteTo interface and returns the number of bytes
// written.
func (h *header) WriteTo(w io.Writer) (int64, error) {
	b := new(bytes.Buffer)
	if err := binary.Write(b, binary.LittleEndian, h); err != nil {
		return 0, err
	}
	n, err := w.Write(b.Bytes())
	return int64(n), err
}

// ReadHeader deserialises a header read from the given Reader.
func ReadHeader(r io.Reader) (*header, error) {
	h := new(header)
	return h, binary.Read(r, binary.LittleEndian, h)
}

// Read implements the standard Read interface and will read up to len(b) bytes
// of unencrypted data into b. To do this Read must read an entire encrypted
// message from the underlying Reader and will block until it is able to do so,
// or an error occurs. If it is unable to decrypt/authenticate the message an
// error will be returned.
func (p *SecureReader) Read(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, nil
	}

	// Check to see if there are still remnants of the last message to read.
	if p.buf.Len() > 0 {
		return p.buf.Read(b)
	}

	// Read header from stream.
	h, err := ReadHeader(p.src)
	if err != nil {
		return 0, err
	}

	// Read the encrypted message.
	e := make([]byte, h.Length)
	if _, err := io.ReadFull(p.src, e); err != nil {
		return 0, err
	}

	// Decrypt message into d (which might be b if it has enough space) and
	// check it is authentic. Limit the capacity of b so that the function can't
	// overwrite array indices >= len(b).
	d, auth := box.OpenAfterPrecomputation(b[:0:len(b)], e[:], &h.Nonce, p.key)
	if !auth {
		return 0, fmt.Errorf("message failed authentication")
	}

	// Check to see if the underlying arrays are the same for b & d slices - if
	// they are then we don't need to copy or buffer anything as implicitly
	// len(d) <= len(b).
	if &d[0] != &b[0] {
		copy(b, d)
		if len(d) > len(b) {
			p.buf = bytes.NewReader(d[len(b):])
			return len(b), nil
		}
	}
	return len(d), nil
}

// Write implements the standard Write interface, as such it returns the number
// of bytes successfully written and will only return a value less than len(b)
// if err != nil.
//
// Write will create a header, encrypt b, and write the results to the
// underlying Writer.
func (p *SecureWriter) Write(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, nil
	}

	// Send b in chunks such that we only send at most maxLength bytes at a
	// time.
	for i := 0; i < len(b); i += maxLength {
		l := min(i+maxLength, len(b))
		h, err := newHeader(len(b[i:l]))
		if err != nil {
			return i, err
		}
		e := box.SealAfterPrecomputation(nil, b[i:l], &h.Nonce, p.key)
		if _, err := h.WriteTo(p.dst); err != nil {
			return i, err
		}
		if _, err := p.dst.Write(e); err != nil {
			return i, err
		}
	}
	return len(b), nil
}

// NewSecureReader instantiates a new SecureReader.
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	sr := SecureReader{src: r, key: new([32]byte), buf: bytes.NewReader(nil)}
	box.Precompute(sr.key, pub, priv)
	return &sr
}

// NewSecureWriter instantiates a new SecureWriter.
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	sw := SecureWriter{dst: w, key: new([32]byte)}
	box.Precompute(sw.key, pub, priv)
	return &sw
}

// swapKeys generates a public/private key pair and swaps the public key with
// a corresponding call to swapKeys over the given ReadWriter.
//
// swapKeys returns the private key generated locally and the public key of it's
// counterpart.
func swapKeys(rw io.ReadWriter) (priv, peer *[32]byte, err error) {
	// Always return nil for the keys if there is an error
	defer func() {
		if err != nil {
			priv, peer = nil, nil
		}
	}()
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return
	}

	// Write our public key
	werr := make(chan error)
	go func() {
		_, err := rw.Write(pub[:])
		werr <- err
	}()
	defer func() {
		if err == nil {
			err = <-werr
		}
	}()

	// Read their public key
	peer = new([32]byte)
	_, err = io.ReadFull(rw, peer[:])
	return
}

// Dial generates a private/public key pair, connects to the server, performs
// the handshake and returns a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	priv, pub, err := swapKeys(conn)
	if err != nil {
		return nil, err
	}
	r := NewSecureReader(conn, priv, pub)
	w := NewSecureWriter(conn, priv, pub)
	return struct {
		io.Reader
		io.Writer
		io.Closer
	}{r, w, conn}, nil
}

// Serve starts a secure echo server on the given listener. Non-fatal errors
// such as individual connection failures are written to the default logger.
func Serve(l net.Listener) error {
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go func() {
			defer conn.Close()
			priv, pub, err := swapKeys(conn)
			if err != nil {
				log.Printf("key exchange failed: %v", err)
				return
			}
			r := NewSecureReader(conn, priv, pub)
			w := NewSecureWriter(conn, priv, pub)
			if _, err := io.Copy(w, r); err != nil {
				log.Printf("message forwarding failed: %v", err)
				return
			}
		}()
	}
}

func main() {
	port := flag.Int("l", 0, "Listen mode. Specify port")
	flag.Parse()

	// Server mode
	if *port != 0 {
		l, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
		if err != nil {
			log.Fatal(err)
		}
		defer l.Close()
		log.Fatal(Serve(l))
	}

	// Client mode
	if len(os.Args) != 3 {
		log.Fatalf("Usage: %s <port> <message>", os.Args[0])
	}
	conn, err := Dial("localhost:" + os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	if _, err := conn.Write([]byte(os.Args[2])); err != nil {
		log.Fatal(err)
	}
	buf := make([]byte, len(os.Args[2]))
	n, err := conn.Read(buf)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", buf[:n])
}
