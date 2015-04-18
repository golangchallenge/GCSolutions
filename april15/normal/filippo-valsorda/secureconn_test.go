package main

import (
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"testing"

	"golang.org/x/crypto/nacl/box"
)

func TestReadWriterPing(t *testing.T) {
	// IMPORTANT: I had to change this test, because the provided version,
	// the one using the same private key for both peers, is actually testing
	// for a security vulnerability. If such a test passes it means that the
	// peer is not performing any check against being served the traffic it
	// generated himself (see: TestReplayAttack)
	//
	// priv, pub := &[32]byte{'p', 'r', 'i', 'v'}, &[32]byte{'p', 'u', 'b'}
	//
	// r, w := io.Pipe()
	// secureR := NewSecureReader(r, priv, pub)
	// secureW := NewSecureWriter(w, priv, pub)

	pubAlice, privAlice, errA := box.GenerateKey(rand.Reader)
	pubBob, privBob, errB := box.GenerateKey(rand.Reader)
	if errA != nil || errB != nil {
		t.Fatal(errA, errB)
	}

	r, w := io.Pipe()
	secureW := NewSecureWriter(w, privAlice, pubBob)
	secureR := NewSecureReader(r, privBob, pubAlice)

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

func TestReadWriter2Write(t *testing.T) {
	pubAlice, privAlice, errA := box.GenerateKey(rand.Reader)
	pubBob, privBob, errB := box.GenerateKey(rand.Reader)
	if errA != nil || errB != nil {
		t.Fatal(errA, errB)
	}

	r, w := io.Pipe()
	secureW := NewSecureWriter(w, privAlice, pubBob)
	secureR := NewSecureReader(r, privBob, pubAlice)

	// Encrypt hello world
	go func() {
		fmt.Fprintf(secureW, "hello")
		fmt.Fprintf(secureW, " world\n")
		w.Close()
	}()

	// Decrypt message
	buf := make([]byte, len("hello world\n"))
	if _, err := io.ReadFull(secureR, buf); err != nil {
		t.Fatal(err)
	}

	// Make sure we have hello world back
	if res := string(buf); res != "hello world\n" {
		t.Fatalf("Unexpected result: %s != %s", res, "hello world")
	}
}

func TestReadWriter2Read(t *testing.T) {
	pubAlice, privAlice, errA := box.GenerateKey(rand.Reader)
	pubBob, privBob, errB := box.GenerateKey(rand.Reader)
	if errA != nil || errB != nil {
		t.Fatal(errA, errB)
	}

	r, w := io.Pipe()
	secureW := NewSecureWriter(w, privAlice, pubBob)
	secureR := NewSecureReader(r, privBob, pubAlice)

	// Encrypt hello world
	go func() {
		fmt.Fprintf(secureW, "h")
		fmt.Fprintf(secureW, "ello world\n")
		w.Close()
	}()

	// Decrypt message
	buf := make([]byte, len("hello world\n"))
	if _, err := io.ReadFull(secureR, buf[:len("hello")]); err != nil {
		t.Fatal(err)
	}
	if _, err := io.ReadFull(secureR, buf[len("hello"):]); err != nil {
		t.Fatal(err)
	}

	// Make sure we have hello world back
	if res := string(buf); res != "hello world\n" {
		t.Fatalf("Unexpected result: %s != %s", res, "hello world")
	}
}

func TestReadWriterEOF(t *testing.T) {
	pubAlice, privAlice, errA := box.GenerateKey(rand.Reader)
	pubBob, privBob, errB := box.GenerateKey(rand.Reader)
	if errA != nil || errB != nil {
		t.Fatal(errA, errB)
	}

	r, w := io.Pipe()
	secureW := NewSecureWriter(w, privAlice, pubBob)
	secureR := NewSecureReader(r, privBob, pubAlice)

	// Encrypt hello world
	go func() {
		fmt.Fprintf(secureW, "hello world\n")
		w.Close()
	}()

	// Decrypt message
	buf := make([]byte, len("hello world\n"))
	if _, err := io.ReadFull(secureR, buf); err != nil {
		t.Fatal(err)
	}

	if n, err := secureR.Read(buf); err != io.EOF || n != 0 {
		t.Fatal(err)
	}
}

func TestReadWriter0LenMsg(t *testing.T) {
	pubAlice, privAlice, errA := box.GenerateKey(rand.Reader)
	pubBob, privBob, errB := box.GenerateKey(rand.Reader)
	if errA != nil || errB != nil {
		t.Fatal(errA, errB)
	}

	r, w := io.Pipe()
	secureW := NewSecureWriter(w, privAlice, pubBob)
	secureR := NewSecureReader(r, privBob, pubAlice)

	// Encrypt hello world
	go func() {
		fmt.Fprintf(secureW, "hello")
		fmt.Fprintf(secureW, "")
		fmt.Fprintf(secureW, " world\n")
		w.Close()
	}()

	// Decrypt message
	buf := make([]byte, len("hello world\n"))
	if _, err := io.ReadFull(secureR, buf); err != nil {
		t.Fatal(err)
	}

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
