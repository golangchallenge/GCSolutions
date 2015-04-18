package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"testing"
	"time"
)

var priv, pub = &[32]byte{'p', 'r', 'i', 'v'}, &[32]byte{'p', 'u', 'b'}

func TestReadWriterPing(t *testing.T) {

	r, w := io.Pipe()
	defer w.Close()
	secureR := NewSecureReader(r, priv, pub)
	secureW := NewSecureWriter(w, priv, pub)

	// Encrypt hello world
	go fmt.Fprintf(secureW, "hello world\n")

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

func TestSecureEchoServer(t *testing.T) {
	// Create a random listener
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	// Start the server
	go Serve(l)

	conn, err := Dial(l.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	time.Sleep(time.Second)
	expected := "hello world\n"
	if _, err := fmt.Fprintf(conn, expected); err != nil {
		t.Fatal(err)
	}

	buf := make([]byte, 2048)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatal(err)
	}

	if got := string(buf[:n]); got != expected {
		t.Fatalf("Unexpected result:\nGot:\t\t%s\nExpected:\t%s\n", got, expected)
	}
}

func TestSecureServe(t *testing.T) {
	// Create a random listener
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	// Start the server
	go Serve(l)

	conn, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	unexpected := "hello world\n"
	if _, err := fmt.Fprintf(conn, unexpected); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 2048)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(buf[:n]); got == unexpected {
		t.Fatalf("Unexpected result:\nGot raw data instead of serialized key")
	}
}

func TestSecureDial(t *testing.T) {
	// Create a random listener
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	// Start the server
	go func(l net.Listener) {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				key := [32]byte{}
				c.Write(key[:])
				buf := make([]byte, 2048)
				n, err := c.Read(buf)
				if err != nil {
					t.Fatal(err)
				}
				if got := string(buf[:n]); got == "hello world\n" {
					t.Fatal("Unexpected result. Got raw data instead of encrypted")
				}
			}(conn)
		}
	}(l)

	conn, err := Dial(l.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	expected := "hello world\n"
	if _, err := fmt.Fprintf(conn, expected); err != nil {
		t.Fatal(err)
	}
}

// Reads one byte each time read is called.
// This is common in network environments when data is streamed as packets arrive,
// not necessarily all at once.
type slowReader struct {
	input []byte
	idx   int
}

func (s *slowReader) Read(p []byte) (int, error) {
	if s.idx >= len(s.input) {
		return 0, io.EOF
	}
	p[0] = s.input[s.idx]
	s.idx++
	return 1, nil
}

func TestPartialRead(t *testing.T) {
	buf := bytes.Buffer{}
	secureW := NewSecureWriter(&buf, priv, pub)

	fmt.Fprintf(secureW, "hello world\n")
	reader := slowReader{buf.Bytes(), 0}
	secureR := NewSecureReader(&reader, priv, pub)
	output := make([]byte, 1000)
	n, err := secureR.Read(output)
	if err != nil {
		t.Fatal(err)
	}
	s := string(output[:n])
	if s != "hello world\n" {
		t.Fatalf("%v != hello world\n", s)
	}
}

// My implementation reads an entire message at once. Message format is [payloadSize(4 bytes) | nonce | payload]
// Any reader must provide sufficient buffer space for the entire payload or an error will occur.
func TestInsufficientBufferRead(t *testing.T) {
	buf := bytes.Buffer{}
	secureW := NewSecureWriter(&buf, priv, pub)
	fmt.Fprintf(secureW, "abcdefg")
	secureR := NewSecureReader(&buf, priv, pub)
	suppliedBuffer := make([]byte, 4)
	_, err := secureR.Read(suppliedBuffer)
	if err == nil {
		t.Fatal("Expected error for insufficient buffer, but got none.")
	}
}

func TestBarelySufficientBufferRead(t *testing.T) {
	buf := bytes.Buffer{}
	secureW := NewSecureWriter(&buf, priv, pub)
	fmt.Fprintf(secureW, "abcdefg")
	secureR := NewSecureReader(&buf, priv, pub)
	suppliedBuffer := make([]byte, 7)
	_, err := secureR.Read(suppliedBuffer)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}
