package drum

import (
	"bytes"
	"testing"
)

func TestStepsError(t *testing.T) {
	b := []byte{0x01, 0x0e, 0x00, 0x00}
	b = append(b, b...)
	b = append(b, b...)

	s, err := readTrackSteps(bytes.NewReader(b))

	if err == nil {
		t.Errorf("Expected error, got steps: %v", s)
	}
}

func TestReadTempo(t *testing.T) {
	b := []byte{0xcd, 0xcc, 0xc4, 0x42}
	r, err := readTempo(bytes.NewReader(b))
	if err != nil || r != 98.4 {
		t.Error(r, err)
	}
}

func TestReadHeader(t *testing.T) {
	b := []byte("SPLICE\x00\x00\x00\x00\x00\x00")
	r, err := readHeader(bytes.NewReader(b))

	if err != nil || r != "SPLICE" {
		t.Error(r, err)
	}
}

func TestReadVersion(t *testing.T) {
	b := []byte("0.808-alpha")
	b = append(b, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}...)
	r, err := readVersion(bytes.NewReader(b))

	if err != nil || r != "0.808-alpha" {
		t.Error(r, err)
	}

}

func TestReadSize(t *testing.T) {
	b := []byte{0x00, 0xC5}
	r, err := readSize(bytes.NewReader(b))

	if err != nil || r != 197 {
		t.Error(r, err)
	}
}

func TestReadTrackID(t *testing.T) {
	b := []byte{0x28, 0x00}
	r, err := readTrackID(bytes.NewReader(b))

	if err != nil || r != 40 {
		t.Error(r, err)
	}
}

func TestReadTrackName(t *testing.T) {
	a := []byte{0, 0, 0, 4}
	b := []byte("Kick")
	c, err := readTrackNameLength(bytes.NewReader(a))
	s, err := readTrackName(bytes.NewReader(b), c)

	if err != nil || s != "Kick" || c != 4 {
		t.Error("name:", s, "err:", err, "count:", c)
	}
}

func TestReadTrackSteps(t *testing.T) {
	b := []byte{0x01, 0x00, 0x00, 0x00}
	b = append(b, b...)
	b = append(b, b...)

	s, err := readTrackSteps(bytes.NewReader(b))

	if err != nil || s[0] != true || s[4] != true || s[9] != false {
		t.Error(s, err)
	}
}
