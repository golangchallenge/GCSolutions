package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"path"
	"strings"
	"testing"
)

func TestBadPath(t *testing.T) {
	_, err := DecodeFile("something_that_you_cant_find")
	if err == nil {
		t.Error("expected test to fail opening inexisting file")
	}
}

func TestBadPattern(t *testing.T) {
	cases := []struct {
		input io.Reader
		err   string
	}{
		{
			strings.NewReader(""),
			"parse splice tag: EOF",
		},
		{
			strings.NewReader("GOPHER"),
			"wrong splice tag: [71 79 80 72 69 82]",
		},
		{
			strings.NewReader("SPLICEtrash"),
			"parse pattern length: unexpected EOF",
		},
		{
			dumpAndRead(struct {
				tag [6]byte
				n   uint64
			}{
				spliceTag,
				1000,
			}),
			"parse version: EOF",
		},
		{
			dumpAndRead(struct {
				tag     [6]byte
				n       uint64
				version [32]byte
			}{
				spliceTag,
				0,
				[32]byte{'g', 'o'},
			}),
			"parse version: reading more than expected",
		},
		{
			dumpAndRead(struct {
				tag     [6]byte
				n       uint64
				version [32]byte
			}{
				spliceTag,
				1000,
				[32]byte{'g', 'o'},
			}),
			"parse tempo: EOF",
		},
		{
			dumpAndRead(struct {
				tag     [6]byte
				n       uint64
				version [32]byte
				tempo   float32
			}{
				spliceTag,
				1000,
				[32]byte{'g', 'o'},
				120,
			}),
			"parse track id: EOF",
		},
	}
	for _, c := range cases {
		_, err := Decode(c.input)
		checkError(t, err, c.err)
	}
}

func TestBadTrack(t *testing.T) {
	cases := []struct {
		input io.Reader
		err   string
	}{
		{strings.NewReader(""), "parse track id: EOF"},
		{dumpAndRead(uint8(0)), "parse track name length: EOF"},
		{dumpAndRead(struct {
			id uint8
			n  int32
		}{0, 42}), "parse track name: EOF"},
		{dumpAndRead(struct {
			id   uint8
			n    int32
			name [2]byte
		}{0, 2, [2]byte{'g', 'o'}}), "parse track 0 steps: EOF"},
	}

	for _, c := range cases {
		_, err := decodeTrack(c.input)
		checkError(t, err, c.err)
	}

}

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
			t.Errorf("%s wasn't decoded as expect.\nGot:\n%s\nExpected:\n%s",
				exp.path, decoded, exp.output)
		}
	}
}

func checkError(t *testing.T, got error, expected string) {
	if got == nil {
		if expected != "" {
			t.Errorf("expected test to fail with message %q", expected)
		}
	} else {
		if expected == "" {
			t.Errorf("unexpected error %q", got)
		} else if got.Error() != expected {
			t.Errorf("expected test to fail with message %q; failed with %q", expected, got)
		}
	}
}

func dumpAndRead(v interface{}) io.Reader {
	w := new(bytes.Buffer)
	if err := binary.Write(w, binary.LittleEndian, v); err != nil {
		panic(err)
	}
	return w
}
