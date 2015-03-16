package drum

import (
	"fmt"
	"path"
	"testing"
)

func TestReplaceSteps(t *testing.T) {
	tData := struct {
		path    string
		output  string
		replace [16]byte
	}{
		"pattern_1.splice",
		`Saved with HW Version: 0.808-alpha
Tempo: 120
(0) kick	|x--x|x--x|x--x|x--x|
(1) snare	|----|x---|----|x---|
(2) clap	|----|x-x-|----|----|
(3) hh-open	|x--x|x--x|x--x|x--x|
(4) hh-close	|x---|x---|----|x--x|
(5) cowbell	|----|----|--x-|----|
`,
		[16]byte{1, 0, 0, 1, 1, 0, 0, 1, 1, 0, 0, 1, 1, 0, 0, 1},
	}

	decoded, err := DecodeFile(path.Join("fixtures", tData.path))
	if err != nil {
		t.Fatalf("something went wrong decoding %s - %v", tData.path, err)
	}

	decoded.ReplaceSteps(0, tData.replace)
	decoded.ReplaceSteps(3, tData.replace)

	if fmt.Sprint(decoded) != tData.output {
		t.Logf("decoded:\n%#v\n", fmt.Sprint(decoded))
		t.Logf("expected:\n%#v\n", tData.output)
		t.Fatalf("%s steps weren't replaced as expected.\nGot:\n%s\nExpected:\n%s",
			tData.path, decoded, tData.output)
	}

}

func TestAddSteps(t *testing.T) {
	tData := struct {
		path   string
		output string
		add    [16]byte
	}{
		"pattern_1.splice",
		`Saved with HW Version: 0.808-alpha
Tempo: 120
(0) kick	|x--x|x--x|x--x|x--x|
(1) snare	|----|x---|----|x---|
(2) clap	|----|x-x-|----|----|
(3) hh-open	|--xx|--xx|x-xx|--xx|
(4) hh-close	|x---|x---|----|x--x|
(5) cowbell	|----|----|--x-|----|
`,
		[16]byte{0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 1},
	}

	decoded, err := DecodeFile(path.Join("fixtures", tData.path))
	if err != nil {
		t.Fatalf("something went wrong decoding %s - %v", tData.path, err)
	}

	decoded.AddSteps(0, tData.add)
	decoded.AddSteps(3, tData.add)

	if fmt.Sprint(decoded) != tData.output {
		t.Logf("decoded:\n%#v\n", fmt.Sprint(decoded))
		t.Logf("expected:\n%#v\n", tData.output)
		t.Fatalf("%s steps weren't added as expected.\nGot:\n%s\nExpected:\n%s",
			tData.path, decoded, tData.output)
	}

}

func TestRemoveSteps(t *testing.T) {
	tData := struct {
		path   string
		output string
		remove [16]byte
	}{
		"pattern_1.splice",
		`Saved with HW Version: 0.808-alpha
Tempo: 120
(0) kick	|----|x---|----|x---|
(1) snare	|----|x---|----|x---|
(2) clap	|----|x-x-|----|----|
(3) hh-open	|----|--x-|----|--x-|
(4) hh-close	|x---|x---|----|x--x|
(5) cowbell	|----|----|--x-|----|
`,
		[16]byte{1, 1, 1, 1, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0},
	}

	decoded, err := DecodeFile(path.Join("fixtures", tData.path))
	if err != nil {
		t.Fatalf("something went wrong decoding %s - %v", tData.path, err)
	}

	decoded.RemoveSteps(0, tData.remove)
	decoded.RemoveSteps(3, tData.remove)

	if fmt.Sprint(decoded) != tData.output {
		t.Logf("decoded:\n%#v\n", fmt.Sprint(decoded))
		t.Logf("expected:\n%#v\n", tData.output)
		t.Fatalf("%s steps weren't removed as expected.\nGot:\n%s\nExpected:\n%s",
			tData.path, decoded, tData.output)
	}

}
