package secure

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"testing"
)

func encryptThenDecrypt(v []byte) ([]byte, error) {
	priv, pub := &[32]byte{'p', 'r', 'i', 'v'}, &[32]byte{'p', 'u', 'b'}
	buf := new(bytes.Buffer)
	reader := NewReader(buf, priv, pub)
	writer := NewWriter(buf, priv, pub)
	if _, err := writer.Write(v); err != nil {
		return nil, err
	}
	p := make([]byte, len(v))
	_, err := reader.Read(p)
	return p, err
}

func TestSingleByte(t *testing.T) {
	p, err := encryptThenDecrypt([]byte{'a'})
	if err != nil || len(p) != 1 || p[0] != 'a' {
		t.Fatalf("Error of %v or invalid result of %v", err, p)
	}
}

func TestEmpty(t *testing.T) {
	p, err := encryptThenDecrypt([]byte{})
	if err != nil || len(p) != 0 {
		t.Fatalf("Error of %v or invalid result of %v", err, p)
	}
}

func TestAtMaximum(t *testing.T) {
	atMax := make([]byte, maxMessageSize)
	if _, err := rand.Read(atMax); err != nil {
		t.Fatalf("Failed making chunk: %v", err)
	}
	p, err := encryptThenDecrypt(atMax)
	if err != nil {
		t.Fatalf("Error of %v", err)
	} else if !bytes.Equal(p, atMax) {
		t.Fatal("Result not the same: %v, %v", len(p), len(atMax))
	}
}

func TestOverMaximum(t *testing.T) {
	overMax := make([]byte, maxMessageSize+1)
	if _, err := rand.Read(overMax); err != nil {
		t.Fatalf("Failed making chunk: %v", err)
	}
	if _, err := encryptThenDecrypt(overMax); err != ErrTooLarge {
		t.Fatal("Expecting error, was success")
	}
}

func TestMultiWriteAndRead(t *testing.T) {
	priv, pub := &[32]byte{'p', 'r', 'i', 'v'}, &[32]byte{'p', 'u', 'b'}
	buf := new(bytes.Buffer)
	reader := NewReader(buf, priv, pub)
	writer := NewWriter(buf, priv, pub)
	fmt.Fprint(writer, "Test-one\n")
	fmt.Fprint(writer, "Test-two\nTest-")
	fmt.Fprint(writer, "three\nTest-four\n")
	expected := []string{"Test-one", "Test-two", "Test-three", "Test-four"}
	var str string
	for i := 0; i < len(expected); i++ {
		if fmt.Fscanln(reader, &str); str != expected[i] {
			t.Fatalf("Expected %v got %v", expected[i], str)
		}
	}
}

// Simple ReadWriteCloser tests are handled by main_test
