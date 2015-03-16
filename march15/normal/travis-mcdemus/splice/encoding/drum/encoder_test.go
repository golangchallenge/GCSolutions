package drum

import (
	"bytes"
	"io/ioutil"
	"path"
	"testing"
)

func TestEncodeTrack(t *testing.T) {
	expected := []byte{2, 0, 0, 0,
		4, 'c', 'l', 'a', 'p',
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 1, 0, 1,
		0, 0, 1, 1}
	track := Track{ID: 2, Name: "clap",
		Sequence: []byte{1, 0, 0, 0,
			0, 1, 0, 0,
			0, 1, 0, 1,
			0, 0, 1, 1}}
	actual := track.encode()
	if len(expected) != len(actual) {
		t.Fatalf("Expected %v output bytes and got %v", len(expected), len(actual))
	}
	for i, b := range actual {
		if expected[i] != b {
			t.Fatalf("Expected '%v' byte but received '%v' at %v", expected[i], b, i)
		}
	}
}

func TestEncode(t *testing.T) {
	unknownIndexes := map[int]struct{}{13: {}, 45: {}, 46: {}, 49: {}}
	tData := []struct {
		path   string
		backup string
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
	}

	for _, input := range tData {
		b := new(bytes.Buffer)
		e := NewEncoder(b)
		pattern, err := NewPatternFromBackup(input.backup)
		err = e.Encode(*pattern)
		actual := b.Bytes()
		if err != nil {
			t.Fatalf("Something went wrong encoding - %v", err)
		}
		filePath := path.Join("patterns", input.path)
		expected, err := ioutil.ReadFile(filePath)
		if err != nil {
			t.Fatal(err)
		}
		if len(expected) != len(actual) {
			t.Fatalf("Expected %v output bytes and got %v", len(expected), len(actual))
		}
		for i, b := range actual {
			if _, ok := unknownIndexes[i]; ok {
				continue
			}
			if expected[i] != b {
				t.Fatalf("Expected '%v' byte but received '%v' at %v", expected[i], b, i)
			}
		}
	}
}
