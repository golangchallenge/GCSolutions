package drum

import (
	"bytes"
	"io/ioutil"
	"path"
	"testing"
)

func TestEncodeFile(t *testing.T) {

	tData := []struct {
		path    string
		pattern *Pattern
	}{
		{
			"pattern_1.splice",
			&Pattern{
				size:    uint8(0xc5),
				Version: "0.808-alpha",
				Tempo:   float32(120),
				Tracks: []*Track{
					&Track{0, "kick", Steps{1, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0}},
					&Track{1, "snare", Steps{0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0}},
					&Track{2, "clap", Steps{0, 0, 0, 0, 1, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0}},
					&Track{3, "hh-open", Steps{0, 0, 1, 0, 0, 0, 1, 0, 1, 0, 1, 0, 0, 0, 1, 0}},
					&Track{4, "hh-close", Steps{1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 1}},
					&Track{5, "cowbell", Steps{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0}},
				},
			},
		},
		{
			"pattern_2.splice",
			&Pattern{
				size:    uint8(0x8f),
				Version: "0.808-alpha",
				Tempo:   float32(98.4),
				Tracks: []*Track{
					&Track{0, "kick", Steps{1, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0}},
					&Track{1, "snare", Steps{0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0}},
					&Track{3, "hh-open", Steps{0, 0, 1, 0, 0, 0, 1, 0, 1, 0, 1, 0, 0, 0, 1, 0}},
					&Track{5, "cowbell", Steps{0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0}},
				},
			},
		},
	}

	for _, exp := range tData {
		actual, err := exp.pattern.MarshalBinary()
		if err != nil {
			t.Fatalf("something went wrong when doing MarshalBinary: %v", err)
		}

		want, _ := ioutil.ReadFile(path.Join("fixtures", exp.path))
		if !bytes.Equal(want, actual) {
			t.Fatalf("want % x \ngot % x", want, actual)
		}
	}
}
