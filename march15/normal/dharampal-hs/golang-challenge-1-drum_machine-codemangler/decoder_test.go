package drum

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"path"
	"testing"
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
			t.Fatalf("%s wasn't decoded as expected.\nGot:\n%s\nExpected:\n%s",
				exp.path, decoded, exp.output)
		}
	}
}

func TestDecodeFileReturnsNilWithErrorWhenUnableToOpenFile(t *testing.T) {
	pattern, err := DecodeFile("non-existent-file")
	assert.Nil(t, pattern)
	assert.Equal(t, "open non-existent-file: no such file or directory", err.Error())
}

func TestPatternStringRepresentation(t *testing.T) {
	header := Header{version: [32]byte{'1', '0', '.', '2', '4', '-', 'b', 'e', 't', 'a'}, tempo: 120}
	trackOne := Track{id: 220,
		name:  &PascalString{length: 9, text: []byte{'L', 'o', 'w', ' ', 'C', 'o', 'n', 'g', 'a'}},
		steps: [16]uint8{0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00}}

	trackTwo := Track{id: 42,
		name:  &PascalString{length: 5, text: []byte{'C', 'r', 'a', 's', 'h'}},
		steps: [16]uint8{0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x00}}

	pattern := Pattern{header: &header, tracks: []Track{trackOne, trackTwo}}

	expectedStringRepresentation :=
		`Saved with HW Version: 10.24-beta
Tempo: 120
(220) Low Conga	|---x|----|---x|----|
(42) Crash	|--x-|--x-|--x-|--x-|
`
	assert.Equal(t, expectedStringRepresentation, fmt.Sprint(pattern))
}
