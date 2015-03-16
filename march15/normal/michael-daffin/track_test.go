package drum

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"testing"
)

func TestStepString(t *testing.T) {
	tData := []struct {
		step   Step
		output string
	}{
		{
			Step(0),
			"-",
		},
		{
			Step(1),
			"x",
		},
		{
			Step(2),
			"x",
		},
	}

	for _, exp := range tData {
		if fmt.Sprint(exp.step) != exp.output {
			t.Fatalf("Step wasn't converted to string as expect.\nGot:\n%s\nExpected:\n%s",
				exp.step, exp.output)
		}
	}
}

func TestTrackString(t *testing.T) {
	tData := []struct {
		track  Track
		output string
	}{
		{
			Track{0, "test1", [4][4]Step{{0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}}},
			"(0) test1\t|----|----|----|----|",
		},
		{
			Track{1, "test2", [4][4]Step{{1, 1, 1, 1}, {1, 1, 1, 1}, {1, 1, 1, 1}, {1, 1, 1, 1}}},
			"(1) test2\t|xxxx|xxxx|xxxx|xxxx|",
		},
		{
			Track{2, "test3", [4][4]Step{{0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}}},
			"(2) test3\t|----|----|----|----|",
		},
		{
			Track{255, "test test test test test", [4][4]Step{{1, 0, 0, 0}, {0, 1, 0, 0}, {0, 0, 1, 0}, {0, 0, 0, 1}}},
			"(255) test test test test test\t|x---|-x--|--x-|---x|",
		},
		{
			Track{0, "test123 []'#;'", [4][4]Step{{0, 1, 1, 0}, {1, 0, 0, 1}, {1, 0, 1, 0}, {0, 1, 0, 1}}},
			"(0) test123 []'#;'\t|-xx-|x--x|x-x-|-x-x|",
		},
	}

	for _, exp := range tData {
		if fmt.Sprint(exp.track) != exp.output {
			t.Fatalf("Track wasn't converted to string as expect.\nGot:\n%s\nExpected:\n%s",
				exp.track, exp.output)
		}
	}
}

func TestTrackDecode(t *testing.T) {
	tData := []struct {
		track     Track
		encoding  []byte
		bytesRead uint
	}{
		{
			Track{0, "test1", [4][4]Step{{0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}}},
			[]byte{0, 0, 0, 0, 5, 't', 'e', 's', 't', '1', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			26,
		},
		{
			Track{1, "test2", [4][4]Step{{1, 1, 1, 1}, {1, 1, 1, 1}, {1, 1, 1, 1}, {1, 1, 1, 1}}},
			[]byte{1, 0, 0, 0, 5, 't', 'e', 's', 't', '2', 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			26,
		},
		{
			Track{2, "test3", [4][4]Step{{0, 0, 0, 0}, {1, 1, 1, 1}, {0, 0, 0, 0}, {0, 0, 0, 0}}},
			[]byte{2, 0, 0, 0, 5, 't', 'e', 's', 't', '3', 0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0},
			26,
		},
		{
			Track{255, "test test test test test", [4][4]Step{{1, 0, 0, 0}, {0, 1, 0, 0}, {0, 0, 1, 0}, {0, 0, 0, 1}}},
			[]byte{255, 0, 0, 0, 24,
				't', 'e', 's', 't', ' ',
				't', 'e', 's', 't', ' ',
				't', 'e', 's', 't', ' ',
				't', 'e', 's', 't', ' ',
				't', 'e', 's', 't',
				1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1},
			45,
		},
		{
			Track{0, "test123 []'#;'", [4][4]Step{{0, 1, 1, 0}, {1, 0, 0, 1}, {1, 0, 1, 0}, {0, 1, 0, 1}}},
			[]byte{0, 0, 0, 0, 14, 't', 'e', 's', 't', '1', '2', '3', ' ', '[', ']', '\'', '#', ';', '\'',
				0, 1, 1, 0, 1, 0, 0, 1, 1, 0, 1, 0, 0, 1, 0, 1},
			35,
		},
	}
	for _, exp := range tData {
		track := Track{}
		r := bytes.NewBuffer(exp.encoding)
		bytesRead, err := track.Decode(r)
		if err != nil {
			t.Fatalf("something went wrong decoding %s - %v", hex.Dump(exp.encoding), err)
		}

		if bytesRead != exp.bytesRead {
			t.Fatalf("Unexpected number of bytes read: Got %d expected %d", bytesRead, exp.bytesRead)
		}

		if exp.track.ID != track.ID {
			t.Fatalf("Unexpected ID: Got %d expected %d", track.ID, exp.track.ID)
		}

		if exp.track.Title != track.Title {
			t.Fatalf("Unexpected title: Got %s expected %s", track.Title, exp.track.Title)
		}

		if exp.track.Steps != track.Steps {
			got := make([]string, len(track.Steps))
			expected := make([]string, len(track.Steps))

			for index, group := range track.Steps {
				for _, step := range group {
					got[index] += step.String()
				}
			}
			for index, group := range exp.track.Steps {
				for _, step := range group {
					expected[index] += step.String()
				}
			}
			t.Fatalf("Unexpected steps: Got [%s] expected [%s]", strings.Join(got, "|"), strings.Join(expected, "|"))
		}
	}
}

func TestTrackDecodeErrors(t *testing.T) {
	tData := []struct {
		encoding []byte
		err      error
	}{
		{
			[]byte{},
			io.EOF,
		},
		{
			[]byte{0, 0, 0, 0, 0},
			io.EOF,
		},
	}
	for _, exp := range tData {
		track := Track{}
		r := bytes.NewBuffer(exp.encoding)
		_, err := track.Decode(r)
		if err != exp.err {
			t.Fatalf("Expected error not found, got: %v - expected: %v", err, exp.err)
		}
	}
}
