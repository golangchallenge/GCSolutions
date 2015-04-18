package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"sync"
	"testing"
)

func TestReadWriterPing(t *testing.T) {
	priv, pub := &key{'p', 'r', 'i', 'v'}, &key{'p', 'u', 'b'}

	r, w := io.Pipe()
	secureR := NewSecureReader(r, priv, pub)
	secureW := NewSecureWriter(w, priv, pub)

	// Encrypt hello world.
	go func() {
		fmt.Fprintf(secureW, "hello world\n")
		w.Close()
	}()

	// Decrypt message.
	buf := make([]byte, 1024)
	n, err := secureR.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	buf = buf[:n]

	// Make sure we have hello world back.
	if res := string(buf); res != "hello world\n" {
		t.Fatalf("Unexpected result: %s != %s", res, "hello world")
	}
}

func TestSecureWriter(t *testing.T) {
	priv, pub := &key{'p', 'r', 'i', 'v'}, &key{'p', 'u', 'b'}

	r, w := io.Pipe()
	secureW := NewSecureWriter(w, priv, pub)

	// Make sure we are secure
	// Encrypt hello world.
	go func() {
		fmt.Fprintf(secureW, "hello world\n")
		w.Close()
	}()

	// Read from the underlying transport instead of the decoder.
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

	// Make sure we are unique.
	// Encrypt hello world.
	go func() {
		fmt.Fprintf(secureW, "hello world\n")
		w.Close()
	}()

	// Read from the underlying transport instead of the decoder.
	buf2, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	// Make sure we dont' read the plain text message.
	if string(buf) == string(buf2) {
		t.Fatal("Unexpected result. The encrypted message is not unique.")
	}
}

func TestSecureWriteConcurrentUse(t *testing.T) {
	priv, pub := &key{'p', 'r', 'i', 'v'}, &key{'p', 'u', 'b'}

	r, w := io.Pipe()
	secureR := NewSecureReader(r, priv, pub)
	secureW := NewSecureWriter(w, priv, pub)

	var wg sync.WaitGroup
	wg.Add(10)

	for i := 0; i < 10; i++ {
		go func() {
			if _, err := fmt.Fprint(secureW, "Hello"); err != nil {
				t.Fatal(err)
			}
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		w.Close()
	}()

	var count int

	for {
		buf := make([]byte, 1024)
		n, err := secureR.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Fatal(err)
		}
		buf = buf[:n]

		switch string(buf) {
		case "Hello":
			count++
		}
	}

	if count != 10 {
		t.Fatalf("expected hello count to be 10, is %v", count)
	}
}
