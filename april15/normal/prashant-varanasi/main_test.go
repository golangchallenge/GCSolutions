package main

import (
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"sync"
	"testing"

	"golang.org/x/crypto/nacl/box"
)

func TestReadWriterPing(t *testing.T) {
	pub1, priv1, _ := box.GenerateKey(rand.Reader)
	pub2, priv2, _ := box.GenerateKey(rand.Reader)

	r, w := io.Pipe()
	secureR := NewSecureReader(r, priv1, pub2)
	secureW := NewSecureWriter(w, priv2, pub1)

	// Encrypt hello world
	go func() {
		fmt.Fprintf(secureW, "hello world\n")
		w.Close()
	}()

	// Decrypt message
	buf := make([]byte, 1024)
	n, err := secureR.Read(buf)
	if err != nil && err != io.EOF {
		t.Fatal(err)
	}
	buf = buf[:n]

	// Make sure we have hello world back
	if res := string(buf); res != "hello world\n" {
		t.Fatalf("Unexpected result: %s != %s", res, "hello world")
	}
}

func TestReadWriterPingOneByte(t *testing.T) {
	pub1, priv1, _ := box.GenerateKey(rand.Reader)
	pub2, priv2, _ := box.GenerateKey(rand.Reader)

	r, w := io.Pipe()
	secureR := NewSecureReader(r, priv1, pub2)
	secureW := NewSecureWriter(w, priv2, pub1)

	// Encrypt hello world
	go func() {
		str := "hello world\n"
		n, err := secureW.Write([]byte(str))
		if err != nil {
			t.Errorf("write failed: %v", err)
		}
		if n > len(str) {
			t.Errorf("write wrote more bytes than buf: wrote %v bytes but buffer is %v bytes",
				n, len(str))
		}
		w.Close()
	}()

	// Decrypt message
	var data []byte
	for {
		var buf [1]byte
		n, err := secureR.Read(buf[:])
		if n > 0 {
			data = append(data, buf[0])
		}
		if err == io.EOF {
			break
		} else if err != nil {
			t.Fatalf("Read failed: %v", err)
		}
	}
	buf := data

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

	expected := "hello world\n"
	if _, err := fmt.Fprintf(conn, expected); err != nil {
		t.Fatal(err)
	}

	buf := make([]byte, 2048)
	n, err := conn.Read(buf)
	if err != nil && err != io.EOF {
		t.Fatal(err)
	}

	if got := string(buf[:n]); got != expected {
		t.Fatalf("Unexpected result:\nGot:\t\t%s\nExpected:\t%s\n", got, expected)
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
	if err != nil && err != io.EOF {
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

	// Start the server, and avoid closing connections till the test logic is done.
	var done sync.WaitGroup
	done.Add(1)
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
				if err != nil && err != io.EOF {
					t.Fatal(err)
				}
				if got := string(buf[:n]); got == "hello world\n" {
					t.Fatal("Unexpected result. Got raw data instead of encrypted")
				}
				done.Wait()

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
	done.Done()
}
