package drum

import (
	"os"
	"testing"
)

func TestFixedString(t *testing.T) {
	expected := "Go World!"
	b := []byte(expected)

	reader := newByteArrayReader(b)

	got := ""
	reader.FixedString(uint8(len(expected)), &got)
	if expected != got {
		t.Fatalf("Expected: %s Got: %s\n", expected, got)
	}
}

func TestVarString(t *testing.T) {
	expected := "Thank You Matt!"
	b := []byte{byte(len(expected))}
	b = append(b, expected...)

	reader := newByteArrayReader(b)

	got := ""
	reader.VarString(&got)
	if expected != got {
		t.Fatalf("Expected: %s Got: %s\n", expected, got)
	}
}

func TestOpeningInvalidFile(t *testing.T) {
	reader := newFileReader("nonexistantfile.splice")
	if !os.IsNotExist(reader.Err()) {
		t.Fatalf("Expected ErrNotExist.  Got %v\n", reader.Err())
	}
}
