package drum

import (
	"fmt"
	"path"
	"testing"
)

func TestEncodeFile(t *testing.T) {
	tData := []struct {
		pat  *Pattern
		path string
	}{
		// Your everyday Dubstep beat
		{
			pat: &Pattern{
				Version: "GoDrum-0.1.0",
				Tempo:   140,
				Tracks: []*Track{
					{
						ID:    0,
						Name:  "Kick",
						Steps: [16]byte{1, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0},
					},
					{
						ID:    1,
						Name:  "Snare",
						Steps: [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0},
					},
					{
						ID:    2,
						Name:  "HiHat",
						Steps: [16]byte{1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0},
					},
				},
			},
			path: "dubstep_pattern.splice",
		},
		{
			pat: &Pattern{
				Version: "GoDrum-0.1.0",
				Tempo:   110,
				Tracks: []*Track{
					{
						ID:    0,
						Name:  "kick",
						Steps: [16]byte{1, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0},
					},
					{
						ID:    1,
						Name:  "snare",
						Steps: [16]byte{0, 0, 0, 1, 0, 0, 1, 0, 0, 1, 0, 1, 0, 0, 1, 0},
					},
					{
						ID:    2,
						Name:  "hi-hat",
						Steps: [16]byte{0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 1, 0},
					},
				},
			},
			path: "mombathon.splice",
		},
	}

	for _, exp := range tData {
		pth := path.Join("fixtures", exp.path)
		if err := EncodeFile(pth, exp.pat); err != nil {
			t.Fatalf("Something went wrong while encoding %s: %v", pth, err)
		}
		decoded, err := DecodeFile(pth)
		if err != nil {
			t.Fatalf("something went wrong while decoding %s: %v", pth, err)
		}
		if fmt.Sprint(decoded) != fmt.Sprint(exp.pat) {
			t.Logf("decoded:\n%#v\n", fmt.Sprint(decoded))
			t.Logf("expected:\n%#v\n", fmt.Sprint(exp.pat))
			t.Fatalf("%s wasn't decoded as expect.\nGot:\n%s\nExpected:\n%s",
				pth, decoded, exp.pat)
		}
	}
}
