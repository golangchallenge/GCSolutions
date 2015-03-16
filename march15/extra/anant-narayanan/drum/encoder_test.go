package drum

import (
	"fmt"
	"os"
	"path"
	"testing"
)

// Simple encoding test, decoding/encoding/encoding the same file
// should result in the same output.
func TestEncodeFile(t *testing.T) {
	tmpFile := "tmp.splice"
	for _, exp := range testData {
		fpath := path.Join("fixtures", exp.path)
		decoded, err := DecodeFile(fpath)
		if err != nil {
			t.Fatalf("something went wrong decoding %s - %v", exp.path, err)
		}
		err = EncodeFile(decoded, tmpFile)
		if err != nil {
			t.Fatalf("something went wrong encoding %s - %v", exp.path, err)
		}
		decoded, err = DecodeFile(tmpFile)
		if err != nil {
			t.Fatalf("something went wrong decoding the encoded file %s - %v", exp.path, err)
		}
		if fmt.Sprint(decoded) != exp.output {
			t.Logf("decoded:\n%#v\n", fmt.Sprint(decoded))
			t.Logf("expected:\n%#v\n", exp.output)
			t.Fatalf("%s wasn't encoded then decoded as expected.\nGot:\n%s\nExpected:\n%s",
				exp.path, decoded, exp.output)
		}
		err = os.Remove(tmpFile)
		if err != nil {
			t.Fatalf("couldn't remove temporary file %s - %v\n", tmpFile, err)
		}
	}
}

// Load a file, modify it and save it back. Check if we get the
// expected output.
func TestEncodeFileChanged(t *testing.T) {
	modified := "pattern_2-morebells.splice"
	modifiedSplice := `Saved with HW Version: 1.00-alpha
Tempo: 98.4
(0) kick	|x---|----|x---|----|
(1) snare	|----|x---|----|x---|
(3) hh-open	|--x-|--x-|x-x-|--x-|
(5) cowbell	|x---|x-x-|x---|x-x-|
`

	decoded, err := DecodeFile(path.Join("fixtures", "pattern_2.splice"))
	if err != nil {
		t.Fatalf("couldn't decode pattern_2 - %v\n", err)
	}
	decoded.Version = "1.00-alpha"
	if decoded.Tracks[3].Name != "cowbell" {
		t.Fatalf("expected 3rd track to be cowbell, found %s\n", decoded.Tracks[3].Name)
	}
	decoded.Tracks[3].Steps = []byte{
		0x1, 0x0, 0x0, 0x0,
		0x1, 0x0, 0x1, 0x0,
		0x1, 0x0, 0x0, 0x0,
		0x1, 0x0, 0x1, 0x0,
	}
	err = EncodeFile(decoded, modified)
	if err != nil {
		t.Fatalf("couldn't encode modified pattern: %v\n", err)
	}
	decoded, err = DecodeFile(modified)
	if err != nil {
		t.Fatalf("couldn't decode modified pattern: %v\n", err)
	}
	if fmt.Sprint(decoded) != modifiedSplice {
		t.Fatalf("decoded splice didn't match expected modification: %v\n", err)
	}
	err = os.Remove(modified)
	if err != nil {
		t.Fatalf("couldn't remove temporary file %s - %v\n", modified, err)
	}
}
