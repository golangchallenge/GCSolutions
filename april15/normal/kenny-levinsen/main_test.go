package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"testing"
)

func TestReadWriterPing(t *testing.T) {
	priv, pub := &[32]byte{'p', 'r', 'i', 'v'}, &[32]byte{'p', 'u', 'b'}

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
	priv, pub := &[32]byte{'p', 'r', 'i', 'v'}, &[32]byte{'p', 'u', 'b'}

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

	var b bytes.Buffer
	secureW = NewSecureWriter(&b, priv, pub)

	// Test for length failure
	// Attempt encryption of message over MaxLength
	longMsg := make([]byte, MaxLength+1)
	n, err := secureW.Write(longMsg)

	if !(n == 0 && err == ErrOutputTooLong) {
		t.Fatalf("Unexpected result. The write did not fail.\nGot:\t\terr: %s, n: %d\nExpected:\terr: %s, n: %d\n", err, n, ErrOutputTooLong, 0)
	}
}

func TestSecureReader(t *testing.T) {
	priv, pub := &[32]byte{'p', 'r', 'i', 'v'}, &[32]byte{'p', 'u', 'b'}

	var b bytes.Buffer
	secureR := NewSecureReader(&b, priv, pub)

	// Test for too small slice failure
	// Send a message instructing to read longer than size the of reading slice
	longMsg := make([]byte, LengthBytes)
	binary.BigEndian.PutUint32(longMsg[0:LengthBytes], MaxLength)

	_, err := b.Write(longMsg)
	if err != nil {
		t.Fatal(err)
	}

	buf := make([]byte, 1)
	n, err := secureR.Read(buf)
	if !(n == 0 && err == ErrSliceTooSmall) {
		t.Fatalf("Unexpected result. The read did not fail.\nGot:\t\terr: %s, n: %d\nExpected:\terr: %s, n: %d\n",
			err, n, ErrSliceTooSmall, 0)
	}

	// Test for immediate return on 0 length slice
	// Requests a read to a slice of 0 length.
	buf = make([]byte, 0)
	n, err = secureR.Read(buf)
	if !(n == 0 && err == nil) {
		t.Fatalf("Unexpected result. The read did return.\nGot:\t\terr: %s, n: %d\nExpected:\terr: %s, n: %d\n",
			err, n, nil, 0)
	}

	b.Reset()
	secureR = NewSecureReader(&b, priv, pub)

	// Test for length failure
	// Send a message instructing to read longer than MaxLength
	longMsg = make([]byte, LengthBytes)
	binary.BigEndian.PutUint32(longMsg[0:LengthBytes], MaxEncryptedLength+1)

	_, err = b.Write(longMsg)
	if err != nil {
		t.Fatal(err)
	}

	buf = make([]byte, 2048)
	n, err = secureR.Read(buf)
	if !(n == 0 && err == ErrInputTooLong) {
		t.Fatalf("Unexpected result. The read did not fail.\nGot:\t\terr: %s, n: %d\nExpected:\terr: %s, n: %d\n",
			err, n, ErrInputTooLong, 0)
	}

	b.Reset()
	secureR = NewSecureReader(&b, priv, pub)

	// Test for verification faliure
	// Send a message with all 0 nonce and message to trip a failure
	brokenMsg := make([]byte, LengthBytes+NonceBytes+8)
	binary.BigEndian.PutUint32(brokenMsg[0:LengthBytes], LengthBytes+NonceBytes+8)

	_, err = b.Write(brokenMsg)
	if err != nil {
		t.Fatal(err)
	}

	buf = make([]byte, 2048)
	n, err = secureR.Read(buf)
	if !(n == 0 && err == ErrVerification) {
		t.Fatalf("Unexpected result. The read did not fail.\nGot:\t\terr: %s, n: %d\nExpected:\terr: %s, n: %d\n",
			err, n, ErrVerification, 0)
	}
}

func TestSecureEchoServer(t *testing.T) {
	// Create a random listener
	l, err := net.Listen("tcp", "127.0.0.1:0")
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

	// We loop here to test sending multiple packages in the same session
	for i := 0; i < 20; i++ {
		expected := fmt.Sprintf("hello world: %b\n", i)
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
}

func TestSecureServe(t *testing.T) {
	// Create a random listener
	l, err := net.Listen("tcp", "127.0.0.1:0")
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
	l, err := net.Listen("tcp", "127.0.0.1:0")
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
