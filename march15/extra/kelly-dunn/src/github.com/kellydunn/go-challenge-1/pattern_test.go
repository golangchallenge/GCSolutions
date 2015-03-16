package drum

import (
	"bytes"
	"testing"
)

var TestVersion = []byte{
	0x30, 0x2e, 0x38, 0x30,
	0x38, 0x2d, 0x61, 0x6c,
	0x70, 0x68, 0x61, 0x00,
	0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00,
}

var TestTempo = []byte{
	0x00, 0x00, 0xf0, 0x42,
	0x00, 0x00, 0x00, 0x00,
}

func TestReadPatternVersion(t *testing.T) {
	p := &Pattern{}
	expected := "0.808-alpha"
	reader := bytes.NewReader(TestVersion)

	read, err := readPatternVersion(reader, p)
	if err != nil {
		t.Errorf("Error reading track name %v", err)
	}

	if p.Version != expected {
		t.Errorf("Mismatched pattern version.  Expected: %s.  Actual: %s", expected, p.Version)
	}

	if read != VersionSize {
		t.Errorf("Mismatched bytes read. Expected %d.  Actual: %d", VersionSize, read)
	}
}

func TestReadPatternTempo(t *testing.T) {
	p := &Pattern{}
	expected := float32(120.0)
	reader := bytes.NewReader(TestTempo)

	read, err := readPatternTempo(reader, p)
	if err != nil {
		t.Errorf("Error reading track name %v", err)
	}

	if p.Tempo != expected {
		t.Errorf("Mismatched pattern tempo.  Expected: %f.  Actual: %f", expected, p.Tempo)
	}

	if read != TempoSize {
		t.Errorf("Mismatched bytes read. Expected %d.  Actual: %d", TempoSize, read)
	}
}
