package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"testing"
)

var priv, pub = &[32]byte{'p', 'r', 'i', 'v'}, &[32]byte{'p', 'u', 'b'}

func TestReadWriterPing(t *testing.T) {
	r, w := io.Pipe()
	secureR := NewSecureReader(r, priv, pub)
	secureW := NewSecureWriter(w, priv, pub)

	// Encrypt hello world
	go func() {
		fmt.Fprintf(secureW, "hello world\n")
		w.Close()
	}()

	// Decrypt message
	buf := make([]byte, 1024)
	n, err := secureR.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	buf = buf[:n]

	// Make sure we have hello world back
	if res := string(buf); res != "hello world\n" {
		t.Fatalf("Unexpected result: %s != %s", res, "hello world")
	}
}

func TestSecureWriter(t *testing.T) {
	r, w := io.Pipe()
	secureW := NewSecureWriter(w, priv, pub)

	// Make sure we are secure
	// Encrypt hello world
	go func() {
		fmt.Fprintf(secureW, "hello world\n")
		w.Close()
	}()

	// Read from the underlying transport instead of the decoder
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	// Make sure we dont' read the plain text message.
	if res := string(buf); res == "hello world\n" {
		t.Fatal("Unexpected result. The message is not encrypted.")
	}

	r, w = io.Pipe()
	secureW = NewSecureWriter(w, priv, pub)

	// Make sure we are unique
	// Encrypt hello world
	go func() {
		fmt.Fprintf(secureW, "hello world\n")
		w.Close()
	}()

	// Read from the underlying transport instead of the decoder
	buf2, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	// Make sure we dont' read the plain text message.
	if string(buf) == string(buf2) {
		t.Fatal("Unexpected result. The encrypted message is not unique.")
	}
}

func TestSecureWriterCheckingMessageLength(t *testing.T) {
	secureW := NewSecureWriter(ioutil.Discard, priv, pub)
	var message []byte
	_, err := secureW.Write(message)
	if err == nil {
		t.Fatal("Error expected when writing empty message")
	}
	message = make([]byte, 32*1024+1)
	_, err = secureW.Write(message)
	if err == nil {
		t.Fatal("Error expected when writing too big message")
	}
	message = []byte{'a', 'b', 'c'}
	n, err := secureW.Write(message)
	if err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Fatalf("Unexecpted number of bytes written: %d", n)
	}
}

// stubWriter implements io.Writer and returns specified length and error.
type stubWriter struct {
	n   int
	err error
}

func (w *stubWriter) Write([]byte) (int, error) {
	return w.n, w.err
}

func TestSecureWriterShortWriting(t *testing.T) {
	expectedErr := errors.New("test error")
	stubW := stubWriter{0, expectedErr}
	secureW := NewSecureWriter(&stubW, priv, pub)
	_, err := secureW.Write([]byte{1})
	if err != expectedErr {
		t.Fatalf("Unexpected error: %v", err)
	}
	stubW.n = 2
	stubW.err = nil
	_, err = secureW.Write([]byte{1, 2, 3})
	if err == nil {
		t.Fatal("Error execpted for short writing")
	}
}

// frameBuffer implements io.ReadWriter. It saves the frame on writing and
// allows reading partially and circularly.
type frameBuffer struct {
	data []byte
	i, n int
	err  error
}

func (r *frameBuffer) Write(p []byte) (int, error) {
	r.data = make([]byte, len(p))
	copy(r.data, p)
	return len(p), r.err
}

func (r *frameBuffer) Read(p []byte) (int, error) {
	n := r.n
	if n == 0 || len(p) < n {
		n = len(p)
	}
	copy(p, r.data[r.i:r.i+n])
	r.i += n
	if r.i >= len(r.data) {
		r.i = 0
	}
	return n, r.err
}

func TestSecureReaderPartialReading(t *testing.T) {
	// frameBuffer will only return 2 bytes for each reading
	frame := frameBuffer{n: 2}

	secureW := NewSecureWriter(&frame, priv, pub)
	secureR := NewSecureReader(&frame, priv, pub)
	_, err := secureW.Write([]byte("abc"))
	if err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 7)
	n, err := secureR.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	if n != 3 || string(buf[:n]) != "abc" {
		t.Fatalf("Unexpected message read: %d - %#v", n, buf)
	}

	frame.i = 0
	buf = make([]byte, 2)
	secureR.(*secureReader).sequence = 0
	_, err = secureR.Read(buf)
	if err != io.ErrShortBuffer {
		t.Fatalf("Unexpected error: %#v", err)
	}
}

func TestSecureReaderWithInvalidLength(t *testing.T) {
	frame := frameBuffer{}

	secureW := NewSecureWriter(&frame, priv, pub)
	secureR := NewSecureReader(&frame, priv, pub)
	_, err := secureW.Write([]byte("1234567"))
	if err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 16)
	binary.BigEndian.PutUint32(frame.data[lengthPos:], 9999999)
	_, err = secureR.Read(buf)
	if err != ErrInvalidHeader {
		t.Fatalf("Unexpected error: %#v", err)
	}
}

func TestSecureReaderWithInvalidNonce(t *testing.T) {
	frame := frameBuffer{}

	secureW := NewSecureWriter(&frame, priv, pub)
	secureR := NewSecureReader(&frame, priv, pub)
	_, err := secureW.Write([]byte("xyz"))
	if err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 16)
	_, err = secureR.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	_, err = secureR.Read(buf)
	if err != ErrInvalidHeader {
		t.Fatalf("Unexpected error: %#v", err)
	}
}

func BenchmarkSecureWriter(b *testing.B) {
	w := NewSecureWriter(ioutil.Discard, priv, pub)
	for i := 0; i < b.N; i++ {
		if _, err := w.Write([]byte("helloworld\n")); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSecureReader(b *testing.B) {
	frame := &frameBuffer{}
	w := NewSecureWriter(frame, priv, pub)

	if _, err := w.Write([]byte("helloworld\n")); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()

	r := NewSecureReader(frame, priv, pub)
	rbuf := make([]byte, maxMessageSize)
	for i := 0; i < b.N; i++ {
		// Reset sequence to ignore replay check.
		r.(*secureReader).sequence = 0
		if _, err := r.Read(rbuf); err != nil {
			b.Fatal(err)
		}
	}
}
