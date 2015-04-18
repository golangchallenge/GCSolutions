package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"math/rand"
	"reflect"
	"testing"
)

var priv, pub = &[32]byte{'p', 'r', 'i', 'v'}, &[32]byte{'p', 'u', 'b'}

var rander = rand.New(rand.NewSource(0))

func makeBytes(n int) []byte {
	b := make([]byte, n)
	for i := 0; i < n; i += 8 {
		num := rander.Int63() // int63, good enough
		for j := 0; j < 8; j++ {
			b[i+j] = uint8(num & 0xff)
			num = num >> 8
		}
	}

	return b
}

// fill plainText with data
// we don't use crypto/rand so that we can ensure consistent data between runs
var benchPlainText = makeBytes(32768)

func BenchmarkSecureWrite(b *testing.B) {
	b.StopTimer()
	sw := NewSecureWriter(ioutil.Discard, priv, pub)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		sw.Write(benchPlainText)
	}
}

func BenchmarkSecureRead(b *testing.B) {
	b.StopTimer()
	buf := bytes.NewBuffer(nil)
	sw := NewSecureWriter(buf, priv, pub)
	sw.Write(benchPlainText)
	bufReader := bytes.NewReader(buf.Bytes())
	sr := NewSecureReader(bufReader, priv, pub)
	b.StartTimer()

	readBuf := make([]byte, len(benchPlainText))
	for i := 0; i < b.N; i++ {
		bufReader.Seek(0, 0)
		sr.lastNonceID = 0
		sr.Read(readBuf)
	}
}

func TestSecureReader_interface(t *testing.T) {
	var _ io.Reader = &SecureReader{}
}

func TestSecureWriter_interface(t *testing.T) {
	var _ io.Writer = &SecureWriter{}
}

func TestSecureReader(t *testing.T) {
	sb := bytes.NewBuffer(nil)
	sw := NewSecureWriter(sb, priv, pub)
	sr := NewSecureReader(sb, priv, pub)

	pt := makeBytes(128)
	if _, err := sw.Write(pt); err != nil {
		t.Fatal(err)
	}

	// test a simple 16 byte read
	buf1 := make([]byte, 16)
	n, err := sr.Read(buf1)
	if err != nil {
		t.Fatalf("Error during read: %s", err)
	}
	if n != len(buf1) {
		t.Fatalf("Incorrect read length: %d != %d", n, len(buf1))
	}
	if !reflect.DeepEqual(buf1, pt[0:len(buf1)]) {
		t.Fatalf("Incorrect result: %q != %q", buf1, pt[:len(buf1)])
	}

	// test that the next read picks up where the previous left off
	buf2Start := len(buf1)
	buf2End := buf2Start + 16
	buf2 := make([]byte, buf2End-buf2Start)
	n, err = sr.Read(buf2)
	if err != nil {
		t.Fatalf("Error during read: %s", err)
	}
	if n != len(buf2) {
		t.Fatalf("Incorrect read length: %d != %d", n, len(buf2))
	}
	if !reflect.DeepEqual(buf2, pt[buf2Start:buf2End]) {
		t.Fatalf("Incorrect result: %q != %q", buf2, pt[buf2Start:buf2End])
	}

	// test that we receive the rest of pt
	buf3Start := len(buf1) + len(buf2)
	buf3End := len(pt)
	// add a few extra bytes to make sure nothing fishy happens with a short read
	buf3 := make([]byte, buf3End-buf3Start+12)
	n, err = sr.Read(buf3)
	if err != nil {
		t.Fatalf("Error during read: %s", err)
	}
	if n != len(buf3)-12 {
		t.Fatalf("Incorrect read length: %d != %d", n, len(buf3)-12)
	}
	if !reflect.DeepEqual(buf3[:n], pt[buf3Start:]) {
		t.Fatalf("Incorrect result: %q != %q", buf3[:n], pt[buf3Start:])
	}

	// test that we receive a second message
	pt = makeBytes(128)
	if _, err := sw.Write(pt); err != nil {
		t.Fatal(err)
	}
	buf4 := make([]byte, len(pt))
	n, err = sr.Read(buf4)
	if err != nil {
		t.Fatalf("Error during read: %s", err)
	}
	if n != len(pt) {
		t.Fatalf("Incorrect read length: %d != %d", n, len(pt))
	}
	if !reflect.DeepEqual(buf4[:n], pt[:]) {
		t.Fatalf("Incorrect result: %#v != %#v", buf4[:n], pt[:])
	}
}

// test reading into a byte array with len() smaller than cap() and make sure
// we don't fill the byte array past len().
func TestSecureReader_overflow(t *testing.T) {
	sb := bytes.NewBuffer(nil)
	sw := NewSecureWriter(sb, priv, pub)
	sr := NewSecureReader(sb, priv, pub)

	pt := makeBytes(32)
	if _, err := sw.Write(pt); err != nil {
		t.Fatal(err)
	}

	buf := make([]byte, 64)
	n, err := sr.Read(buf[:16])
	if err != nil {
		t.Fatalf("Error during read: %s", err)
	}
	if n != 16 {
		t.Fatalf("Incorrect read length: %d != %d", n, 16)
	}
	if !reflect.DeepEqual(buf[:16], pt[:16]) {
		t.Fatalf("Incorrect result: %q != %q", buf[:16], pt[:16])
	}
	for i := 16; i < len(buf); i++ {
		if buf[i] != 0x00 {
			t.Fatalf("Buffer was overflowed. buf[%d] != 0x00", i)
		}
	}
}

// test that reader returns an error if the nonce is reused
func TestSecureReader_reuseNonce(t *testing.T) {
	sb := bytes.NewBuffer(nil)
	sw := NewSecureWriter(sb, priv, pub)
	sr := NewSecureReader(sb, priv, pub)

	pt := makeBytes(32)
	if _, err := sw.Write(pt); err != nil {
		t.Fatal(err)
	}

	ct := make([]byte, sb.Len())
	copy(ct, sb.Bytes())

	buf := make([]byte, len(pt))
	_, err := sr.Read(buf)
	if err != nil {
		t.Fatalf("Error during read: %s", err)
	}

	// put the same cipher text back on the wire and ensure Read() errors
	_, err = sb.Write(ct)
	if err != nil {
		t.Fatalf("Error during write: %s", err)
	}

	_, err = sr.Read(buf)
	if err == nil {
		t.Fatalf("Expected error during read, but none returned")
	}
	if err.Error() != "invalid nonce" {
		t.Fatalf("Wrong error received. %q != \"invalid nonce\"", err.Error())
	}
}
