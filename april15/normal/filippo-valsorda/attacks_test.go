package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"io/ioutil"
	"testing"

	"golang.org/x/crypto/nacl/box"
)

// TestReorderAttack tries to perform an attack where messages from Alice to Bob
// are reordered on the wire by an active MitM attacker. This would mean
// sequence numbers are not properly working.
func TestReorderAttack(t *testing.T) {
	pubA, privA, err := box.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	pubB, privB, err := box.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	buf := &bytes.Buffer{}

	// Instantiate Alice's writer
	wr := NewSecureWriter(buf, privA, pubB)

	// Write and capture two messages
	if _, err := wr.Write([]byte("message1")); err != nil {
		t.Fatal(err)
	}
	message1 := buf.String()
	buf.Reset()
	if _, err := wr.Write([]byte("message2")); err != nil {
		t.Fatal(err)
	}
	message2 := buf.String()

	t.Log("\n" + hex.Dump([]byte(message1)))
	t.Log("\n" + hex.Dump([]byte(message2)))

	// Check that Bob's reader can read the messages correctly
	buf = bytes.NewBufferString(message1 + message2)
	rd := NewSecureReader(buf, privB, pubA)
	res, err := ioutil.ReadAll(rd)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(res, []byte("message1message2")) {
		t.Fatalf("Bob read the wrong message: %s", res)
	}

	// Check that Bob's reader refuses to read the messages out of order
	buf = bytes.NewBufferString(message2 + message1)
	rd = NewSecureReader(buf, privB, pubA)
	res, err = ioutil.ReadAll(rd)
	if err == nil {
		t.Fatalf("Bob read the out-of-order messages: %s", res)
	}
	t.Logf("The out-of-order error is: %v", err)
}

// TestReplayAttack tries to perform an attack where messages from Alice to Bob
// are sent back to Alice by an active MitM attacker. This would mean that
// parties have no way to distinguish the other, like with even/odd sequence
// numbers for high/low public keys.
func TestReplayAttack(t *testing.T) {
	pubA, privA, err := box.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	pubB, privB, err := box.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	buf := &bytes.Buffer{}

	// Instantiate Alice's writer
	wr := NewSecureWriter(buf, privA, pubB)

	// Write and capture two messages
	if _, err := wr.Write([]byte("message1")); err != nil {
		t.Fatal(err)
	}
	message1 := buf.String()
	buf.Reset()
	if _, err := wr.Write([]byte("message2")); err != nil {
		t.Fatal(err)
	}
	message2 := buf.String()

	t.Log("\n" + hex.Dump([]byte(message1)))
	t.Log("\n" + hex.Dump([]byte(message2)))

	// Check that Bob's reader can read the messages correctly
	buf = bytes.NewBufferString(message1 + message2)
	rd := NewSecureReader(buf, privB, pubA)
	res, err := ioutil.ReadAll(rd)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(res, []byte("message1message2")) {
		t.Fatalf("Bob read the wrong message: %s", res)
	}

	// Check that Alice's reader refuses to read the messages
	buf = bytes.NewBufferString(message2 + message1)
	rd = NewSecureReader(buf, privA, pubB)
	res, err = ioutil.ReadAll(rd)
	if err == nil {
		t.Fatalf("Alice read the messages she sent: %s", res)
	}
	t.Logf("The replayed messages error is: %v", err)
}
