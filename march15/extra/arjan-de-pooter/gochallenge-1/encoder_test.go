package drum

import (
	"bytes"
	"testing"
)

func TestEncodePattern(t *testing.T) {
	pattern := &Pattern{
		Version: "0.915-dev",
		Tempo:   100.6,
		Tracks: []*Track{
			&Track{
				ID:   1,
				Name: "kick",
				Steps: [16]bool{
					false, false, true, true,
					true, true, false, false,
					false, false, false, true,
					true, true, true, true,
				},
			},
			&Track{
				ID:   329,
				Name: "HiHat",
				Steps: [16]bool{
					false, false, false, true,
					false, false, false, true,
					false, false, false, true,
					false, false, false, true,
				},
			},
		},
	}
	var encoded bytes.Buffer
	_, err := encoded.ReadFrom(pattern.Encode())
	if err != nil {
		t.Fatalf("something went wrong encoding %v", err)
	}

	expected := []byte("SPLICE" +
		"\x00\x00\x00\x00\x00\x00\x00\x57" +
		"0.915-dev\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00" +
		"\x33\x33\xc9\x42" +
		"\x01\x00\x00\x00\x04kick\x00\x00\x01\x01\x01\x01\x00\x00\x00\x00\x00\x01\x01\x01\x01\x01" +
		"\x49\x01\x00\x00\x05HiHat\x00\x00\x00\x01\x00\x00\x00\x01\x00\x00\x00\x01\x00\x00\x00\x01")

	t.Logf("Expected: %s (%d)", expected, len(expected))
	t.Logf("Got: %s (%d)", encoded.Bytes(), len(encoded.Bytes()))
	if !bytes.Equal(encoded.Bytes(), expected) {
		t.Fatalf("Encoded data not valid")
	}
}
