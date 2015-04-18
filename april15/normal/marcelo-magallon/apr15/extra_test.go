package main

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"testing"

	"golang.org/x/crypto/nacl/box"
)

const testPlaintext = "hello world\n"

func makeTestKeys() (*[32]byte, *[32]byte) {
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
	return priv, pub
}

func wrapTestReaderAndWriter(t *testing.T, r io.Reader, w io.Writer) (io.Reader, io.Writer, func(), func()) {
	readerPriv, readerPub := makeTestKeys()
	writerPriv, writerPub := makeTestKeys()

	secureR := NewSecureReader(r, readerPriv, writerPub)
	if secureR == nil {
		t.Fatalf("Failed to create a SecureReader")
	}
	secureW := NewSecureWriter(w, writerPriv, readerPub)
	if secureW == nil {
		t.Fatalf("Failed to create a SecureWriter")
	}
	rCloser := func() {
		if rc, ok := r.(io.ReadCloser); ok {
			rc.Close()
		}
	}
	wCloser := func() {
		if wc, ok := w.(io.WriteCloser); ok {
			wc.Close()
		}
	}
	return secureR, secureW, rCloser, wCloser
}

func makeTestReaderAndWriter(t *testing.T) (io.Reader, io.Writer, func(), func()) {
	r, w := io.Pipe()
	return wrapTestReaderAndWriter(t, r, w)
}

func TestWriteClosedReader(t *testing.T) {
	_, w, rCloser, wCloser := makeTestReaderAndWriter(t)
	rCloser()
	defer wCloser()

	n, err := w.Write([]byte(testPlaintext))
	if n != 0 || err == nil {
		t.Fatal("Unexpected result. Writer should not be able to write.")
	}
}

func TestWriteClosedWriter(t *testing.T) {
	_, w, rCloser, wCloser := makeTestReaderAndWriter(t)
	defer rCloser()
	wCloser()

	n, err := w.Write([]byte(testPlaintext))
	if n != 0 || err == nil {
		t.Fatal("Unexpected result. Writer should not be able to write.")
	}
}

func TestReadClosedReader(t *testing.T) {
	r, _, rCloser, wCloser := makeTestReaderAndWriter(t)
	rCloser()
	defer wCloser()

	_, err := ioutil.ReadAll(r)
	if err == nil {
		t.Fatal("Unexpected result. Reader should not be able to read.")
	}
}

func TestReadClosedWriter(t *testing.T) {
	r, _, rCloser, wCloser := makeTestReaderAndWriter(t)
	defer rCloser()
	wCloser()

	buf, err := ioutil.ReadAll(r)
	if err != io.ErrUnexpectedEOF {
		t.Fatal("Unexpected result. Reader should report io.ErrUnexpectedEOF.")
	}
	if len(buf) != 0 {
		t.Fatal("Unexpected result. Buffer should be empty.")
	}
}

func TestShortMessageRead(t *testing.T) {
	r, w := io.Pipe()

	secureR, secureW, rCloser, wCloser := wrapTestReaderAndWriter(t, r, w)
	defer rCloser()
	defer wCloser()

	// Write a full message
	go secureW.Write([]byte(testPlaintext))

	// steal the header + cypher text from the SecureReader
	// XXX: We are abusing knowledge of the internal implementation.
	// We know that since the underlyig io.Pipe writes entire
	// messages, two Read's are necessary.
	s := 0
	msg := make([]byte, 1024)

	n, _ := r.Read(msg[s:])
	s += n

	h := s // this is the header size

	n, _ = r.Read(msg[s:])
	s += n // this is the total message size

	// Write a short message
	go func() {
		w.Write(msg[:(h+s)/2])
		// the close is necessary for the SecureReader's
		// underlying ReadFull to stop trying to read the full
		// message.
		w.Close()
	}()

	// Read the short message using the SecureReader
	buf := make([]byte, h+s)
	if _, err := secureR.Read(buf); err != io.ErrUnexpectedEOF {
		t.Fatalf("Unexpected result: expecting io.ErrUnexpectedEOF, got %v.", err)
	}
}

func TestShortHeaderRead(t *testing.T) {
	buf := bytes.NewBuffer(make([]byte, 0, 1024))

	secureR, secureW, _, _ := wrapTestReaderAndWriter(t, buf, buf)

	// Write a full message
	secureW.Write([]byte(testPlaintext))

	// keep the header + cyphertext
	msg := make([]byte, buf.Len())
	buf.Read(msg)

	// Write back a short message with a truncated header
	buf.Write(msg[:4+8])

	// Read the short message
	if _, err := secureR.Read(msg); err != io.ErrUnexpectedEOF {
		t.Fatalf("Unexpected result: expecting io.ErrUnexpectedEOF, got %v.", err)
	}
}

func TestReaderDecryptionError(t *testing.T) {
	buf := bytes.NewBuffer(make([]byte, 0, 1024))

	secureR, secureW, _, _ := wrapTestReaderAndWriter(t, buf, buf)

	// Write a full message
	secureW.Write([]byte(testPlaintext))

	// keep the header + cypher text
	msg := make([]byte, buf.Len())
	buf.Read(msg)

	// corrupt cyphertext
	i := (headerLen + len(msg)) / 2
	msg[i] = ^msg[i]

	// Write back corrupted message
	buf.Write(msg)

	// Read the corrupted message
	if _, err := secureR.Read(msg); err != ErrDecryptionError {
		t.Fatalf("Unexpected result: expecting ErrDecryptionError, got %v.", err)
	}
}

func TestShortMessage(t *testing.T) {
	buf := bytes.NewBuffer(make([]byte, 0, MsgOverhead))

	secureR, secureW, _, _ := wrapTestReaderAndWriter(t, buf, buf)

	// Write a short message
	switch n, err := secureW.Write([]byte{}); {
	case err != nil:
		t.Fatalf("Unexpected result: expecting no error, got %v.", err)
	case n != 0:
		t.Fatalf("Unexpected result: expecting 0 bytes written, got %d.", n)
	}

	msg := make([]byte, MsgOverhead)

	// Read the short message
	switch n, err := secureR.Read(msg); {
	case err != nil:
		t.Fatalf("Unexpected result: expecting no error, got %v.", err)
	case n != 0:
		t.Fatalf("Unexpected result: expecting 0 bytes read, got %d.", n)
	}
}

func TestLongMessage(t *testing.T) {
	buf := bytes.NewBuffer(make([]byte, 0))

	secureR, secureW, _, _ := wrapTestReaderAndWriter(t, buf, buf)

	msg := make([]byte, MaxMsgLen)

	// Write a long message
	switch n, err := secureW.Write(msg); {
	case err != nil:
		t.Fatalf("Unexpected result: expecting no error, got %v.", err)
	case n != len(msg):
		t.Fatalf("Unexpected result: expecting len(msg) bytes written, got %d.", n)
	}

	// Read the long message
	switch n, err := secureR.Read(msg); {
	case err != nil:
		t.Fatalf("Unexpected result: expecting no error, got %q.", err)
	case n != len(msg):
		t.Fatalf("Unexpected result: expecting MaxMsgLen bytes read, got %d.", n)
	}
}

func TestHugeRead(t *testing.T) {
	header := struct {
		DataLen int32
		Nonce   [nonceLen]byte
	}{
		DataLen: MaxMsgLen + box.Overhead + 1,
	}

	buf := bytes.NewBuffer(make([]byte, 0))

	secureR, _, _, _ := wrapTestReaderAndWriter(t, buf, buf)

	if err := binary.Write(buf, binary.BigEndian, &header); err != nil {
		t.Fatalf("Unexpected result: expecting no error, got %q.", err)
	}

	switch n, err := secureR.Read([]byte{}); {
	case err != ErrMsgTooLarge:
		t.Fatalf("Unexpected result: expecting ErrMsgTooLarge, got %q.", err)
	case n != 0:
		t.Fatalf("Unexpected result: no bytes should have been read, got %d.", n)
	}
}

func TestHugeWrite(t *testing.T) {
	buf := bytes.NewBuffer(make([]byte, 0))

	_, secureW, _, _ := wrapTestReaderAndWriter(t, buf, buf)

	msg := make([]byte, MaxMsgLen+1024)

	// Write a long message
	switch n, err := secureW.Write(msg); {
	case err != ErrMsgTooLarge:
		t.Fatalf("Unexpected result: expecting ErrMsgTooLarge, got %q.", err)
	case n != 0:
		t.Fatalf("Unexpected result: no bytes should have been written, got %d.", n)
	}
}

func TestShortReadBuffer(t *testing.T) {
	buf := bytes.NewBuffer(make([]byte, 0))

	secureR, secureW, _, _ := wrapTestReaderAndWriter(t, buf, buf)

	out := make([]byte, 32)

	// Write a short message
	switch n, err := secureW.Write(out); {
	case err != nil:
		t.Fatalf("Unexpected result: expecting no error, got %v.", err)
	case n != len(out):
		t.Fatalf("Unexpected result: expecting %d bytes written, got %d.", len(out), n)
	}

	in := make([]byte, 16)

	// Read the short message
	switch n, err := secureR.Read(in); {
	case err != io.ErrShortBuffer:
		t.Fatalf("Unexpected result: expecting io.ErrShortBuffer, got %v.", err)
	case n != 0:
		t.Fatalf("Unexpected result: expecting 0 bytes read, got %d.", n)
	}
}

type WriterWrapper struct {
	W io.Writer
	N int
}

func (w *WriterWrapper) Write(p []byte) (int, error) {
	if len(p) > w.N {
		p = p[:w.N]
	}

	n, err := w.W.Write(p)
	w.N -= n
	if err == nil && w.N == 0 {
		err = errors.New("Reached write limit")
	}

	return n, err
}

func TestWrappedWriteError(t *testing.T) {
	buf := bytes.NewBuffer(make([]byte, 0))

	w := &WriterWrapper{W: buf, N: MsgOverhead + 8}

	_, secureW, _, _ := wrapTestReaderAndWriter(t, buf, w)

	out := make([]byte, 2*w.N)

	switch n, err := secureW.Write(out); {
	case err == nil:
		t.Fatalf("Unexpected result: expecting error, got %v.", err)
	case n != 0:
		t.Fatalf("Unexpected result: expecting 0 bytes written, got %d.", n)
	}
}

func TestDialConnectionErrorHandling(t *testing.T) {
	// Create a random listener
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := l.Addr().String()
	// close immediately to cause error in Dial
	l.Close()

	conn, err := Dial(addr)
	if err == nil {
		t.Fatalf("Unexpected result: expecting error, got %v.", err)
		conn.Close()
	}
}

func TestDialHandshakeErrorHandling(t *testing.T) {
	// Create a random listener
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	// Start the server
	go func(l net.Listener) {
		conn, err := l.Accept()
		if err != nil {
			return
		}
		// don't do handshake, simply close the connection
		conn.Close()
	}(l)

	conn, err := Dial(l.Addr().String())
	if err == nil {
		t.Fatalf("Unexpected result: expecting error, got %v.", err)
		conn.Close()
	}
}

func TestServerBadHandshake(t *testing.T) {
	// Create a random listener
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	go Serve(l)

	addr := l.Addr().String()
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("Unexpected result: expecting no error, got %v.", err)
	}
	defer conn.Close()

	// send a bad handshake message
	conn.Write(make([]byte, 32))

	// read response
	msg := make([]byte, 1024)
	var n int
	if n, err = conn.Read(msg); err != nil {
		t.Fatalf("Unexpected result: expecting no error, got %q.", err)
	}
	if !bytes.Equal(msg[:n], badHandshakeResponse) {
		t.Fatalf("Unexpected result: expecting bad handshake, got %q.", msg)
	}
}
