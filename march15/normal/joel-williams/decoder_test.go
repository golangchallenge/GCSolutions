package drum

import (
	"bytes"
	"fmt"
	"os"
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

func TestDecodeHeader(t *testing.T) {
	tData := []struct {
		path        string
		output      header
		expectError error
	}{
		{
			"pattern_1.splice",
			header{
				SpliceDecl: [6]byte{'S', 'P', 'L', 'I', 'C', 'E'},
				Version:    [11]byte{'0', '.', '8', '0', '8', '-', 'a', 'l', 'p', 'h', 'a'},
				Tempo:      [4]byte{0x0, 0x0, 0xf0, 0x42},
			},
			nil,
		},
		{
			"pattern_6.splice",
			header{},
			ErrNotSliceFile,
		},
	}
	for _, exp := range tData {
		f, err := os.Open(path.Join("fixtures", exp.path))
		if err != nil {
			t.Fatalf("something went wrong with opening %s - %v", exp.path, err)
		}
		defer f.Close()
		decoded, err := decodeHeader(f)
		if err != nil {
			if err == exp.expectError {
				continue
			}
			t.Fatalf("something went wrong with decoding the header for %s - %v", exp.path, err)
		}
		if *decoded != exp.output {
			t.Logf("decoded:\n%#v\n", *decoded)
			t.Logf("expected:\n%#v\n", exp.output)
			t.Fatalf("%s wasn't decoded as expected.\nGot:\n%v\nExpected:\n%v",
				exp.path, *decoded, exp.output)
		}
	}
}

func TestDecodeTempo(t *testing.T) {
	tData := []struct {
		input  [4]byte
		output float32
	}{
		{
			[4]byte{0x0, 0x0, 0xf0, 0x42},
			float32(120),
		},
		{
			[4]byte{0xcd, 0xcc, 0xc4, 0x42},
			float32(98.4),
		},
	}
	for _, exp := range tData {
		out := decodeTempo(exp.input)
		if out != exp.output {
			t.Fatalf("%v wasn't decoded as expected.\nGot:%f\nExpected:%f\n", exp.input, out, exp.output)
		}
	}

}

func TestDecodeTrack(t *testing.T) {
	tData := []struct {
		input  []byte
		output *Track
	}{
		{
			[]byte{0x0, 0x0, 0x0, 0x0, 0x4, 0x6b, 0x69, 0x63, 0x6b, 0x1, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0},
			&Track{
				ID:    0,
				Name:  "kick",
				Steps: [16]Step{true, false, false, false, true, false, false, false, true, false, false, false, true, false, false, false},
			},
		},
		{
			[]byte{0x1, 0x0, 0x0, 0x0, 0x5, 0x73, 0x6e, 0x61, 0x72, 0x65, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0},
			&Track{
				ID:    1,
				Name:  "snare",
				Steps: [16]Step{false, false, false, false, true, false, false, false, false, false, false, false, true, false, false, false},
			},
		},
		{
			[]byte{0x2, 0x5, 0x73, 0x6e, 0x61, 0x72, 0x65, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0},
			nil,
		},
	}
	for _, exp := range tData {
		r := bytes.NewReader(exp.input)
		out := decodeTrack(r)
		if exp.output != out && *out != *exp.output {
			t.Fatalf("%v wasn't decoded as expected.\nGot:%v\nExpected:%v\n", exp.input, out, exp.output)
		}
	}
}
