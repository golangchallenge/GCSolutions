package drum

import (
	"testing"
)

var TestPattern = &Pattern{
	Version: "0.808-alpha",
	Tempo:   123.1,
	Tracks: []*Track{
		&Track{
			ID:   0,
			Name: "kick",
			StepSequence: StepSequence{
				Steps: []byte{
					0, 0, 0, 0,
					0, 0, 0, 0,
					0, 0, 0, 0,
					0, 0, 0, 0,
				},
			},
		},
	},
}

// Tests that a valid pattern can be encoded into a splice file.
func TestEncodePattern(t *testing.T) {
	err := EncodePattern(TestPattern, "fixtures/test-encoded.splice")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// An integration test that asserts that a Encoded file
// Can also be successfully Decoded, and has the same string
// Representation as the original pattern.
func TestEncodeDecode(t *testing.T) {

	err := EncodePattern(TestPattern, "fixtures/test-encoded.splice")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	decoded, err := DecodeFile("fixtures/test-encoded.splice")
	if err != nil {
		t.Errorf("Unexpected Error testing a full integration of decoding and encoding a splice file, %v", err)
	}

	if TestPattern.String() != decoded.String() {
		t.Errorf("Decoded string from newly encoded file does not match.  \nExpected: '%v' \nActual: '%v'", TestPattern.String(), []byte(decoded.String()))
	}
}
