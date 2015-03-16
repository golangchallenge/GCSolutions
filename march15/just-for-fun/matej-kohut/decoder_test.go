package drum

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
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

func TestEncode(t *testing.T) {
	tData := []string{
		"pattern_1.splice",
		"pattern_2.splice",
		"pattern_3.splice",
		"pattern_4.splice",
		"pattern_5.splice",
	}

	for _, filePath := range tData {
		fileData, err := ioutil.ReadFile(path.Join("fixtures", filePath))
		if err != nil {
			t.Fatalf("something went wrong reading %s - %v", filePath, err)
		}
		p := &Pattern{}
		err = p.Read(bytes.NewBuffer(fileData))
		if err != nil {
			t.Fatalf("something went wrong decoding %s - %v", filePath, err)
		}

		patternBuffer := &bytes.Buffer{}
		err = p.Write(patternBuffer)
		if err != nil {
			t.Fatalf("something went wrong writing %s - %v", filePath, err)
		}

		if bytes.Compare(fileData, patternBuffer.Bytes()) != 0 {
			t.Logf("encoded:\n%#v\n", patternBuffer.Bytes())
			t.Logf("expected:\n%#v\n", fileData)
			t.Fatalf("Data from read & write are not the same for %s", filePath)
		}
	}
}

func TestEncodeFile(t *testing.T) {
	tData := []string{
		"pattern_1.splice",
		"pattern_2.splice",
		"pattern_3.splice",
		"pattern_4.splice",
		"pattern_5.splice",
	}

	for _, filePath := range tData {
		fileData, err := ioutil.ReadFile(path.Join("fixtures", filePath))
		if err != nil {
			t.Fatalf("something went wrong reading %s - %v", filePath, err)
		}
		p := &Pattern{}
		err = p.Read(bytes.NewBuffer(fileData))
		if err != nil {
			t.Fatalf("something went wrong decoding %s - %v", filePath, err)
		}
		newFilePath := filePath + ".tmp"
		err = p.EncodeToFile(path.Join("fixtures", newFilePath))
		if err != nil {
			t.Fatalf("something went wrong writing %s - %v", newFilePath, err)
		}

		newFileData, err := ioutil.ReadFile(path.Join("fixtures", newFilePath))
		if err != nil {
			t.Fatalf("something went wrong reading %s - %v", newFilePath, err)
		}

		if bytes.Compare(fileData, newFileData) != 0 {
			t.Logf("saved:\n%#v\n", newFileData)
			t.Logf("expected:\n%#v\n", fileData)
			t.Fatalf("Data from read & write are not the same for %s", filePath)
		}

		os.Remove(path.Join("fixtures", newFilePath))
	}
}

func TestCowbellTrackChange(t *testing.T) {
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
(5) cowbell	|--x-|----|--x-|----|
`,
		},
	}
	for _, exp := range tData {
		decoded, err := DecodeFile(path.Join("fixtures", exp.path))
		if err != nil {
			t.Fatalf("something went wrong decoding %s - %v", exp.path, err)
		}
		decoded.Instruments[5].Steps[2] = 1
		if fmt.Sprint(decoded) != exp.output {
			t.Logf("changed:\n%#v\n", fmt.Sprint(decoded))
			t.Logf("expected:\n%#v\n", exp.output)
			t.Fatalf("%s wasn't changed as expect.\nGot:\n%s\nExpected:\n%s",
				exp.path, decoded, exp.output)
		}
	}
}
