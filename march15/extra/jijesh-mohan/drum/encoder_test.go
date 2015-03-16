package drum

import (
	"fmt"
	"path"
	"testing"
)

func TestEncodeFile(t *testing.T) {
	output := `Saved with HW Version: 0.808-alpha
Tempo: 98.4
(0) kick	|x---|----|x---|----|
(1) snare	|----|x---|----|x---|
(3) hh-open	|--x-|--x-|x-x-|--x-|
(5) cowbell	|x---|x-x-|x---|x-x-|
`
	pattern, err := DecodeFile(path.Join("fixtures", "pattern_2.splice"))
	if err != nil {
		t.Fatalf("something went wrong decoding %v", err)
	}
	track := pattern.Tracks[3]
	track.Steps = [16]byte{1, 0, 0, 0, 1, 0, 1, 0, 1, 0, 0, 0, 1, 0, 1, 0}
	pattern.Tracks[3] = track

	err = EncodeFile(pattern, path.Join("fixtures", "pattern_2-morebells.splice"))

	if err != nil {
		t.Fatalf("something went wrong decoding %v", err)
	}

	decoded, err := DecodeFile(path.Join("fixtures", "pattern_2-morebells.splice"))
	if err != nil {
		t.Fatalf("something went wrong decoding %v", err)
	}

	if fmt.Sprint(decoded) != output {
		t.Logf("decoded:\n%#v\n", fmt.Sprint(decoded))
		t.Logf("expected:\n%#v\n", output)
		t.Fatalf("%s wasn't encoded as expect.\nGot:\n%s\nExpected:\n%s",
			"pattern_2_modified.splice", decoded, output)
	}

}
