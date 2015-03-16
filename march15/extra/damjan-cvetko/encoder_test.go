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
(5) cowbell	|x---|x-x-|x---|x-x-|
`,
		},
		{"pattern_2.splice",
			`Saved with HW Version: 0.808-alpha
Tempo: 98.4
(0) kick	|x---|----|x---|----|
(1) snare	|----|x---|----|x---|
(3) hh-open	|--x-|--x-|x-x-|--x-|
(5) cowbell	|x---|x-x-|x---|x-x-|
`,
		},
	}
	for _, exp := range tData {
		decoded, err := DecodeFile(path.Join("fixtures", exp.path))
		if err != nil {
			t.Fatalf("something went wrong decoding %s - %v", exp.path, err)
		}
		// change file
		decoded.AddTrack(Track{5, "cowbell", [16]bool{true, false, false, false, true, false, true, false, true, false, false, false, true, false, true, false}})
		// write it
		err = decoded.EncodeFile(path.Join("fixtures", "tmp.splice"))
		if err != nil {
			t.Fatalf("something went wrong writing changed tile file %s - %v", exp.path, err)
		}
		// read it again
		decoded, err = DecodeFile(path.Join("fixtures", "tmp.splice"))
		if err != nil {
			t.Fatalf("something went wrong decoding2 %s - %v", exp.path, err)
		}
		if fmt.Sprint(decoded) != exp.output {
			t.Logf("decoded:\n%#v\n", fmt.Sprint(decoded))
			t.Logf("expected:\n%#v\n", exp.output)
			t.Fatalf("%s wasn't decoded as expect.\nGot:\n%s\nExpected:\n%s",
				exp.path, decoded, exp.output)
		}
	}
}
