package drum

import (
	"bytes"
	"fmt"
	"path"
	"strings"
	"testing"
)

func fromHex(c byte) byte {
	switch {
	case '0' <= c && c <= '9':
		return c - '0'
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10
	}
	panic("invalid hex")
}

func fromHexStr(hexStr string) []byte {
	bStrs := strings.Split(hexStr, " ")

	res := make([]byte, len(bStrs))

	for i, v := range bStrs {
		res[i] = (fromHex(v[0]) << 4) | fromHex(v[1])
	}
	return res
}

func TestDecodeTrack(t *testing.T) {
	buf := make([]byte, 1024)

	tData := []struct {
		hexStr string
		id     int
		name   string
		steps  string
	}{
		{
			"63 00 00 00 07 4D 61 72 61 63 61 73 01 00 01 00 01 00 01 00 01 00 01 00 01 00 01 00",
			99,
			"Maracas",
			"|x-x-|x-x-|x-x-|x-x-|",
		},
		{
			"FF 00 00 00 09 4C 6F 77 20 43 6F 6E 67 61 00 00 00 00 01 00 00 00 00 00 00 00 01 00 00 00",
			255,
			"Low Conga",
			"|----|x---|----|x---|",
		},
		{
			"03 00 00 00 07 68 68 2D 6F 70 65 6E 00 00 01 00 00 00 01 00 01 00 01 00 00 00 01 00",
			3,
			"hh-open",
			"|--x-|--x-|x-x-|--x-|",
		},
	}

	for _, exp := range tData {
		trackEncoding := fromHexStr(exp.hexStr)
		r := bytes.NewReader(trackEncoding)
		track, err := decodeTrack(r, buf)
		if err != nil {
			t.Errorf("failed to decode track %s: %v", exp.name, err)
		}

		if track.Id != exp.id {
			t.Errorf("track %s: expected id %d, got id %d", exp.name, exp.id, track.Id)
		}
		if track.Name != exp.name {
			t.Errorf("track %s: expected name %s, got name %s", exp.name, exp.name, track.Name)
		}
		if track.Steps.String() != exp.steps {
			t.Errorf("track %s: expected steps %s, got steps %s", exp.name, exp.steps, track.Steps)
		}
	}
}

func TestDecodePatternPreamble(t *testing.T) {
	buf := make([]byte, 1024)

	tData := []struct {
		hexStr  string
		version string
		tempo   float32
	}{
		{
			"00 00 00 00 00 00 00 C5 30 2E 38 30 38 2D 61 6C 70 68 61 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 F0 42",
			"0.808-alpha",
			120,
		},
		{
			"00 00 00 00 00 00 00 8F 30 2E 38 30 38 2D 61 6C 70 68 61 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 CD CC C4 42",
			"0.808-alpha",
			98.4,
		},
		{
			"00 00 00 00 00 00 00 C5 30 2E 38 30 38 2D 61 6C 70 68 61 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 EC 42",
			"0.808-alpha",
			118,
		},
		{
			"00 00 00 00 00 00 00 93 30 2E 39 30 39 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 70 43",
			"0.909",
			240,
		},
		{
			"00 00 00 00 00 00 00 57 30 2E 37 30 38 2D 61 6C 70 68 61 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 C0 79 44",
			"0.708-alpha",
			999,
		},
	}

	for k, exp := range tData {
		preambleEncoding := fromHexStr(exp.hexStr)
		r := bytes.NewReader(preambleEncoding)
		p, _, err := decodePatternPreamble(r, buf)
		if err != nil {
			t.Errorf("failed to decode preamble %d: %v", k, err)
		}

		if p.Version != exp.version {
			t.Errorf("preamble %d: expected version %s, got version %s", k, exp.version, p.Version)
		}
		if p.Tempo != exp.tempo {
			t.Errorf("preamble %d: expected tempo %f, got tempo %f", k, exp.tempo, p.Tempo)
		}
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
			t.Fatalf("%s wasn't decoded as expect.\nGot:\n%s\nExpected:\n%s",
				exp.path, decoded, exp.output)
		}
	}
}
