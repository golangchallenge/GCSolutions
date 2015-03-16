package drum

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path"
	"testing"
)

func TestPatternUnmarshalAndMarshalBinary(t *testing.T) {
	tFiles := []string{
		"pattern_1.splice",
		"pattern_2.splice",
		"pattern_3.splice",
		"pattern_4.splice",
	}

	for _, tFile := range tFiles {
		decodedData, err := ioutil.ReadFile(path.Join("fixtures", tFile))
		if err != nil {
			t.Fatalf("something went wrong decoding %s - %v", tFile, err)
		}

		pattern := &Pattern{}
		if err := pattern.UnmarshalBinary(decodedData); err != nil {
			t.Fatalf("something went wrong unmarshaling %s - %v", tFile, err)
		}

		encodedData, err := pattern.MarshalBinary()
		if err != nil {
			t.Fatalf("something went wrong marshaling %s - %v", tFile, err)
		}

		if bytes.Compare(decodedData, encodedData) != 0 {
			t.Fatalf("%s wasn't encoded correctly", tFile)
		}
	}
}

func TestPatternAppendTrack(t *testing.T) {
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
(55) yeah	|--x-|xx--|-x--|x--x|
`,
		},
		{"pattern_2.splice",
			`Saved with HW Version: 0.808-alpha
Tempo: 98.4
(0) kick	|x---|----|x---|----|
(1) snare	|----|x---|----|x---|
(3) hh-open	|--x-|--x-|x-x-|--x-|
(5) cowbell	|----|----|x---|----|
(55) yeah	|--x-|xx--|-x--|x--x|
`,
		},
	}

	for _, exp := range tData {
		decodedData, err := ioutil.ReadFile(path.Join("fixtures", exp.path))
		if err != nil {
			t.Fatalf("something went wrong decoding %s - %v", exp.path, err)
		}

		pattern := &Pattern{}
		if err := pattern.UnmarshalBinary(decodedData); err != nil {
			t.Fatalf("something went wrong unmarshaling %s - %v", exp.path, err)
		}

		yeahTrack := Track{
			ID:   55,
			Name: "yeah",
			Steps: [16]bool{
				false, false, true, false,
				true, true, false, false,
				false, true, false, false,
				true, false, false, true,
			},
		}
		pattern.AddTrack(yeahTrack)
		if fmt.Sprint(pattern) != exp.output {
			t.Logf("decoded:\n%#v\n", fmt.Sprint(pattern))
			t.Logf("expected:\n%#v\n", exp.output)
			t.Fatalf("%s wasn't decoded as expect.\nGot:\n%s\nExpected:\n%s",
				exp.path, pattern, exp.output)
		}

	}

}
