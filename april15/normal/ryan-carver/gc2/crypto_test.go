package main

import (
	"bytes"
	"crypto/rand"
	"io"
	"testing"

	"golang.org/x/crypto/nacl/box"
)

func Test_NewKeyPair(t *testing.T) {
	k := NewKeyPair()
	if k == nil {
		t.Fatalf("Got nil, want a key")
	}
	var check = func(k *[32]byte) {
		var a = make([]byte, len(k))
		var b = make([]byte, len(k))
		copy(a, k[:])
		if bytes.Equal(a, b) {
			t.Fatalf("Want non-zero value")
		}
	}
	check(k.priv)
	check(k.pub)
}

func Test_KeyPair_Exchange(t *testing.T) {
	kp := &KeyPair{
		&[32]byte{'a'},
		&[32]byte{'b'},
	}
	peersPub := [32]byte{'c'}

	// Fake a io.ReadWriter
	r := bytes.NewBuffer([]byte{})
	w := bytes.NewBuffer([]byte{})
	rw := struct {
		io.Reader
		io.Writer
	}{r, w}

	// Write the peersPub to the buffer.
	r.Write(peersPub[:])

	kp2, err := kp.Exchange(rw)
	if err != nil {
		t.Fatalf("Exchange got error %s", err)
	}
	if !bytes.Equal(w.Bytes(), kp.pub[:]) {
		t.Errorf("Send pub key: got %#v, want %#v", w.Bytes(), kp.pub)
	}
	// WRONG
	if !bytes.Equal(kp2.pub[:], kp2.pub[:]) {
		t.Errorf("Recv pub key: got %#v, want %#v", kp2.pub, kp2.pub)
	}
	// WRONG
	if !bytes.Equal(kp2.priv[:], kp2.priv[:]) {
		t.Errorf("Recv priv key: got %#v, want %#v", kp2.priv, kp2.priv)
	}
}

func Test_KeyPair_CommonKey(t *testing.T) {
	kp := &KeyPair{
		&[32]byte{'a'},
		&[32]byte{'b'},
	}
	want := CommonKey(kp.pub, kp.priv)
	got := kp.CommonKey()
	if !bytes.Equal(got[:], want[:]) {
		t.Errorf("Common key got %v, want %v", got, want)
	}
}

func Test_KeyPairDiffieHellmanCommonKey(t *testing.T) {
	kp1 := NewKeyPair()
	kp2 := NewKeyPair()
	var common1 *[32]byte
	var common2 *[32]byte
	r, w := io.Pipe()
	go func() {
		x, _ := kp2.recv(r)
		common2 = x.CommonKey()
	}()
	kp1.send(w)
	go func() {
		x, _ := kp1.recv(r)
		common1 = x.CommonKey()
	}()
	kp2.send(w)
	if common1 == nil || common2 == nil {
		t.Fatalf("Common keys must not be nil")
	}
	if !bytes.Equal(common1[:], common2[:]) {
		t.Errorf("Want equal common keys\na: %v\nb: %v\n", common1, common2)
	}
}

func Test_CommonKey(t *testing.T) {
	aPub, aPriv, _ := box.GenerateKey(rand.Reader)
	bPub, bPriv, _ := box.GenerateKey(rand.Reader)
	aCommon := CommonKey(bPub, aPriv)
	bCommon := CommonKey(aPub, bPriv)
	if !bytes.Equal(aCommon[:], bCommon[:]) {
		t.Errorf("Want equal common keys\na: %v\nb: %v", aCommon, bCommon)
	}
}

func Test_NewNonce(t *testing.T) {
	n, err := NewNonce()
	if err != nil {
		t.Fatalf("NewNonce got error %s", err)
	}

	if got := len(n); got != 24 {
		t.Fatalf("Got %d, want 24", got)
	}

	var a = make([]byte, len(n))
	var b = make([]byte, len(n))
	copy(a, n[:])
	if bytes.Equal(a, b) {
		t.Errorf("Want non-zero value")
	}
}

func Test_NonceFrom(t *testing.T) {
	buf := make([]byte, nonceSize+1)
	copy(buf, "hello")

	n, err := NonceFrom(buf)

	if err != nil {
		t.Fatalf("Got error %s, want no error", err)
	}

	a := make([]byte, len(n))
	copy(a, "hello")
	if !bytes.Equal(a, n[:]) {
		t.Errorf("Got %v, want nonce to have value of buffer", n)
	}
}

func Test_Nonce_NonceFrom_fail(t *testing.T) {
	buf := make([]byte, nonceSize-1)

	n, err := NonceFrom(buf)

	if n != nil {
		t.Errorf("Want no nonce")
	}
	if err == nil {
		t.Errorf("Want error")
	}
}
