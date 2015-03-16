package drum

import (
	"errors"
	"fmt"
	"path"
	"testing"
)

func TestDecodeFile(t *testing.T) {
	tData := []struct {
		path        string
		output      string
		errExpected error
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
			nil,
		},
		{"pattern_2.splice",
			`Saved with HW Version: 0.808-alpha
Tempo: 98.4
(0) kick	|x---|----|x---|----|
(1) snare	|----|x---|----|x---|
(3) hh-open	|--x-|--x-|x-x-|--x-|
(5) cowbell	|----|----|x---|----|
`,
			nil,
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
			nil,
		},
		{"pattern_4.splice",
			`Saved with HW Version: 0.909
Tempo: 240
(0) SubKick	|----|----|----|----|
(1) Kick	|x---|----|x---|----|
(99) Maracas	|x-x-|x-x-|x-x-|x-x-|
(255) Low Conga	|----|x---|----|x---|
`,
			nil,
		},
		{"pattern_5.splice",
			`Saved with HW Version: 0.708-alpha
Tempo: 999
(1) Kick	|x---|----|x---|----|
(2) HiHat	|x-x-|x-x-|x-x-|x-x-|
`,
			nil,
		},
		{"pattern_1_broken_header.splice",
			"",
			errors.New("splice header is corrupted, expected 'SPLICE', got 'SLICE\x00'"),
		},
		{"pattern_1_broken_pattern_len.splice",
			"",
			errors.New("wrong number of bytes read, expected '21', got '197'"),
		},
		{"pattern_1_tracks_with_same_id.splice",
			"",
			errors.New("track with id '5' already exists in pattern"),
		},
	}

	for _, exp := range tData {
		decoded, err := DecodeFile(path.Join("fixtures", exp.path))
		if err != nil {

			if exp.errExpected == nil {
				t.Fatalf("something went wrong decoding %s - %v", exp.path, err)
			}

			if exp.errExpected.Error() != err.Error() {
				t.Fatalf("%s produced non-expected error.\nGot:\n%s\nExpected:\n%s",
					exp.path, []byte(err.Error()), []byte(exp.errExpected.Error()))
			}
			continue
		}

		if fmt.Sprint(decoded) != exp.output {
			t.Logf("decoded:\n%#v\n", fmt.Sprint(decoded))
			t.Logf("expected:\n%#v\n", exp.output)
			t.Fatalf("%s wasn't decoded as expect.\nGot:\n%s\nExpected:\n%s",
				exp.path, decoded, exp.output)
		}
	}
}

func TestMoreSteps(t *testing.T) {
	tData := []struct {
		path    string
		output  string
		trackID uint8
	}{
		{"pattern_1.splice",
			`Saved with HW Version: 0.808-alpha
Tempo: 120
(0) kick	|x---|x---|x---|x---|
(1) snare	|----|x---|----|x---|
(2) clap	|----|x-x-|----|----|
(3) hh-open	|--x-|--x-|x-x-|--x-|
(4) hh-close	|x---|x---|----|x--x|
(5) cowbell	|xxxx|-x-x|xxxx|-x-x|
`,
			5,
		},
		{"pattern_2.splice",
			`Saved with HW Version: 0.808-alpha
Tempo: 98.4
(0) kick	|x---|----|x---|----|
(1) snare	|----|x---|----|x---|
(3) hh-open	|--x-|--x-|x-x-|--x-|
(5) cowbell	|xxxx|-x-x|xxxx|-x-x|
`,
			5,
		},
		{"pattern_3.splice",
			`Saved with HW Version: 0.808-alpha
Tempo: 118
(40) kick	|x---|----|x---|----|
(1) clap	|----|x---|----|x---|
(3) hh-open	|--x-|--x-|x-x-|--x-|
(5) low-tom	|xxxx|-x-x|xxxx|-x-x|
(12) mid-tom	|----|----|x---|----|
(9) hi-tom	|----|----|-x--|----|
`,
			5,
		},
	}

	for _, exp := range tData {
		decoded, err := DecodeFile(path.Join("fixtures", exp.path))
		if err != nil {
			t.Fatalf("something went wrong decoding %s - %v", exp.path, err)
		}

		track, ok := decoded.GetTrack(exp.trackID)
		if !ok {
			t.Fatalf("track with id %d was not found in %s", exp.trackID, exp.path)
		}

		track.SetStep(0, true)
		track.SetStep(1, true)
		track.SetStep(2, true)
		track.SetStep(3, true)

		track.SetStep(4, false)
		track.SetStep(5, true)
		track.SetStep(6, false)
		track.SetStep(7, true)

		track.SetStep(8, true)
		track.SetStep(9, true)
		track.SetStep(10, true)
		track.SetStep(11, true)

		track.SetStep(12, false)
		track.SetStep(13, true)
		track.SetStep(14, false)
		track.SetStep(15, true)

		if fmt.Sprint(decoded) != exp.output {
			t.Logf("decoded:\n%#v\n", fmt.Sprint(decoded))
			t.Logf("expected:\n%#v\n", exp.output)
			t.Fatalf("%s wasn't decoded as expect.\nGot:\n%s\nExpected:\n%s",
				exp.path, decoded, exp.output)
		}
	}
}
