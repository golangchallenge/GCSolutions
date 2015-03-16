package drum

import (
	"bytes"
	"encoding/hex"
	"io/ioutil"
	"path"
	"testing"
)

func TestWriteTrackID(t *testing.T) {
	expected := []byte{0x28}
	var buf bytes.Buffer
	track := Track{ID: 40}

	track.writeID(&buf)
	if !bytes.Equal(buf.Bytes(), expected) {
		t.Error("Track ID incorrect")
	}
}

func TestTrackWriteName(t *testing.T) {
	expected := []byte{0x00, 0x00, 0x00, 0x04, 0x6B, 0x69, 0x63, 0x6B}
	var buf bytes.Buffer
	track := Track{Name: "kick"}

	track.writeName(&buf)
	if !bytes.Equal(buf.Bytes(), expected) {
		t.Error("Track name incorrect")
	}
}

func TestTrackWriteSteps(t *testing.T) {
	expected := []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	var buf bytes.Buffer
	track := Track{Steps: []bool{true, false, false, false, false, false, false, false, true, false, false, false, false, false, false, false}}

	track.writeSteps(&buf)
	if !bytes.Equal(buf.Bytes(), expected) {
		t.Errorf("Track steps incorrect: %v", buf.Bytes())
	}
}

func TestTrackWrite(t *testing.T) {
	expected := []byte{0x28, 0x00, 0x00, 0x00, 0x04, 0x6B, 0x69, 0x63, 0x6B, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	var buf bytes.Buffer
	track := Track{ID: 40, Name: "kick", Steps: []bool{true, false, false, false, false, false, false, false, true, false, false, false, false, false, false, false}}

	track.write(&buf)
	if !bytes.Equal(buf.Bytes(), expected) {
		t.Error("Track write incorrect")
	}
}

func TestDecodeEncodePattern(t *testing.T) {
	tData := []string{
		"pattern_1.splice",
		"pattern_2.splice",
		"pattern_3.splice",
	}

	for _, exp := range tData {
		expected, err := ioutil.ReadFile(path.Join("fixtures", exp))
		decoded := Pattern{}
		err = Unmarshal(expected, &decoded)
		if err != nil {
			t.Fatal(err)
		}

		actual, err := Marshal(decoded)

		if !bytes.Equal(actual, expected) {
			t.Errorf("%s was incorrectly encoded", exp)
			t.Log("Expected\n" + hex.Dump(expected))
			t.Log("Got\n" + hex.Dump(actual))
			t.Logf("First diff at index %#x", firstDiff(actual, expected))
		}

	}

}

func firstDiff(a, b []byte) int {
	for i, j := range a {
		if j != b[i] {
			return i
		}
	}
	return -1
}
