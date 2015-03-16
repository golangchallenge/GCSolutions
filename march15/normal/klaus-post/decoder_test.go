package drum

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"testing"
	"time"
)

func TestDecodeFile(t *testing.T) {
	tData := []struct {
		path   string
		output string
	}{
		{"pattern_1.splice",
			`Saved with HW Version: 0.808-alpha
Tempo: 120
(0) kick	|x---|x---|x---|x---|
(1) snare	|----|x---|----|x---|
(2) clap	|----|x-x-|----|----|
(3) hh-open	|--x-|--x-|x-x-|--x-|
(4) hh-close	|x---|x---|----|x--x|
(5) cowbell	|----|----|--x-|----|
`,
		},
		{"pattern_2.splice",
			`Saved with HW Version: 0.808-alpha
Tempo: 98.4
(0) kick	|x---|----|x---|----|
(1) snare	|----|x---|----|x---|
(3) hh-open	|--x-|--x-|x-x-|--x-|
(5) cowbell	|----|----|x---|----|
`,
		},
		{"pattern_3.splice",
			`Saved with HW Version: 0.808-alpha
Tempo: 118
(40) kick	|x---|----|x---|----|
(1) clap	|----|x---|----|x---|
(3) hh-open	|--x-|--x-|x-x-|--x-|
(5) low-tom	|----|---x|----|----|
(12) mid-tom	|----|----|x---|----|
(9) hi-tom	|----|----|-x--|----|
`,
		},
		{"pattern_4.splice",
			`Saved with HW Version: 0.909
Tempo: 240
(0) SubKick	|----|----|----|----|
(1) Kick	|x---|----|x---|----|
(99) Maracas	|x-x-|x-x-|x-x-|x-x-|
(255) Low Conga	|----|x---|----|x---|
`,
		},
		{"pattern_5.splice",
			`Saved with HW Version: 0.708-alpha
Tempo: 999
(1) Kick	|x---|----|x---|----|
(2) HiHat	|x-x-|x-x-|x-x-|x-x-|
`,
		},
	}

	for _, exp := range tData {
		decoded, err := DecodeFile(path.Join("fixtures", exp.path))
		if err != nil {
			t.Fatalf("something went wrong decoding %s - %v", exp.path, err)
		}

		if fmt.Sprint(decoded) != exp.output {
			t.Logf("decoded:\n%#v\n", fmt.Sprint(decoded))
			t.Logf("expected:\n%#v\n", exp.output)
			t.Fatalf("%s wasn't decoded as expect.\nGot:\n%s\nExpected:\n%s",
				exp.path, decoded, exp.output)
		}
	}
}

// Array of all test files, used in TestFuzzFiles
var allFiles = []string{"pattern_1.splice",
	"pattern_2.splice",
	"pattern_3.splice",
	"pattern_4.splice",
	"pattern_5.splice"}

// TestFuzzFiles will do fuzz testing on the decoder
// to check if error conditions are handled properly.
func TestFuzzFiles(t *testing.T) {
	// We set a fixed seed so we can reproduce errors
	rand.Seed(42)
	for _, name := range allFiles {
		file, err := os.Open(path.Join("fixtures", name))
		if err != nil {
			t.Fatal(err)
		}
		// Read content of the file into a buffer
		buf, err := ioutil.ReadAll(file)
		if err != nil {
			t.Fatal(err)
		}

		// Do 1000 runs per file.
		for i := 0; i < 1000; i++ {
			// Create a temporary buffer and copy original content
			tmpBuf := make([]byte, len(buf))
			copy(tmpBuf, buf)
			for j := 0; j < 10; j++ {
				// Change 10 random bytes
				where := rand.Intn(len(buf))
				what := byte(rand.Intn(255) - 127)
				tmpBuf[where] = what
			}
			// We add a 1 in 10 chance that we will truncate the buffer
			if rand.Intn(10) == 0 {
				tmpBuf = tmpBuf[:rand.Intn(len(tmpBuf))]
			}
			// Attempt to decode the buffer.
			// We expect errors, but not panics
			p, err := DecodeReader(bytes.NewBuffer(tmpBuf))

			// If we should get a valid value, test if it can stringify itself
			// without crashing.
			if err == nil {
				_ = p.String()
			}
		}

	}
}

// In TestDecodeFileNotFound we test if we get the expected *os.PathError
// returned if the a file cannot be found.
func TestDecodeFileNotFound(t *testing.T) {
	_, err := DecodeFile(path.Join("fixtures", "non-existing-file.splice"))
	if err == nil {
		t.Fatalf("Expected a file not found, but got no error")
	}
	switch err.(type) {
	case *os.PathError:
	default:
		t.Fatalf("Expected a *os.PathError, but got no error type %T", err)
	}
}

// TestDecodeFileInfo will test information retrieval functions
func TestDecodeFileInfo(t *testing.T) {
	p, err := DecodeFile(path.Join("fixtures", "pattern_5.splice"))
	if err != nil {
		t.Fatal(err)
	}
	// Check basic info
	if len(p.Tracks) != 2 {
		t.Fatalf("Expected 2 tracks, got %d", len(p.Tracks))
	}
	d := p.Duration()
	// 16 beats at 999bpm
	expectD := 16 * time.Minute / 999
	if d != expectD {
		t.Errorf("Expected duration to be %v, got %v", expectD, d)
	}
	// Check we get the tracks we expect.
	t0 := p.Tracks[0]
	if t0.Len() != 16 {
		t.Errorf("Expected track length was 16, got %d", t0.Len())
	}
	if t0.At(8) != StepPlay || t0.At(8+16) != StepPlay {
		t.Errorf("Expected to find hit at 8, 24")
	}
	if t0.At(1) != StepNothing || t0.At(17) != StepNothing {
		t.Errorf("Expected to find no hit at 1 & 17")
	}
	if t0.At(-1) != StepNothing {
		t.Errorf("Expected to find no hit at -1")
	}
	// Check we get what we expect at beat 0
	ins := p.PlayAt(0)
	if len(ins) != 2 {
		t.Fatalf("Expected to get 2 instruments at beat 0, got %d", len(ins))
	}
	if ins[0].Name != "Kick" || ins[0].ID != 1 {
		t.Errorf("Expected to get instrument 'Kick' with id 1, got %s, %d", ins[0].Name, ins[0].ID)
	}
	if ins[1].Name != "HiHat" || ins[1].ID != 2 {
		t.Errorf("Expected to get instrument 'HiHat' with id 2, got %s, %d", ins[1].Name, ins[1].ID)
	}
	// Beat -1 should always return len(0)
	ins = p.PlayAt(-1)
	if len(ins) != 0 {
		t.Errorf("Expected to get 0 instruments at beat -1, got %d", len(ins))
	}
	// Test we only get 1 at beat 4
	ins = p.PlayAt(4)
	if len(ins) != 1 {
		t.Errorf("Expected to get 1 instruments at beat 4, got %d", len(ins))
	}
}
