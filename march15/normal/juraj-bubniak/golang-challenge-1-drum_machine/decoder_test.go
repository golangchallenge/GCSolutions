package drum

import (
	"fmt"
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
			t.Fatalf("%s wasn't decoded as expect.\nGot:\n%s\nExpected:\n%s",
				exp.path, decoded, exp.output)
		}
	}
}

func TestDecodePatterns(t *testing.T) {
	tData := []struct {
		path    string
		tempo   float32
		version string
		tracks  int
	}{
		{
			path:    "pattern_1.splice",
			tempo:   120.0,
			version: "0.808-alpha",
			tracks:  6,
		},
		{
			path:    "pattern_2.splice",
			tempo:   98.4,
			version: "0.808-alpha",
			tracks:  4,
		},
		{
			path:    "pattern_3.splice",
			tempo:   118.0,
			version: "0.808-alpha",
			tracks:  6,
		},
		{
			path:    "pattern_4.splice",
			tempo:   240.0,
			version: "0.909",
			tracks:  4,
		},
		{
			path:    "pattern_5.splice",
			tempo:   999.0,
			version: "0.708-alpha",
			tracks:  2,
		},
	}

	for _, exp := range tData {
		pattern, err := DecodeFile(path.Join("fixtures", exp.path))
		if err != nil {
			t.Fatalf("something went wrong decoding %s - %v", exp.path, err)
		}

		if exp.tempo != pattern.Tempo {
			t.Errorf("invalid tempo, expected = %.2f, got = %.2f, name = %s", exp.tempo, pattern.Tempo, exp.path)
		}

		if exp.version != pattern.Version {
			t.Errorf("invalid version, expected = %s, got = %s, name = %s", exp.version, pattern.Version, exp.path)
		}

		tracks := len(pattern.Tracks)
		if exp.tracks != tracks {
			t.Errorf("invalid track count, expected = %d, got = %d, name = %s", exp.tracks, tracks, exp.path)
		}
	}
}
