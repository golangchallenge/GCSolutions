package drum

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
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

func TestWriteFile(t *testing.T) {
	inputFileName := "pattern_1.splice"
	inputPath := path.Join("fixtures", inputFileName)

	inputHash := fileSHA256(inputPath)

	decoded, err := DecodeFile(inputPath)
	if err != nil {
		t.Fatalf("something went wrong decoding %s - %v", inputFileName, err)
	}

	outputFileName := inputFileName + "_tmp"
	outputPath := path.Join("fixtures", outputFileName)

	err = decoded.Write(outputPath)
	defer os.Remove(outputPath)

	if err != nil {
		t.Fatal("Error writing", err)
	}

	outputHash := fileSHA256(outputPath)

	if inputHash != outputHash {
		t.Fatal("Output file does not match input file")
	}
}

func TestModifyFile(t *testing.T) {
	tData := []struct {
		path   string
		output string
	}{
		{"pattern_1.splice",
			`Saved with HW Version: 1.0-golang
Tempo: 240
(0) kick	|x---|x---|x---|x---|
(1) snare	|----|x---|----|x---|
(2) clap	|----|x-x-|----|----|
(3) hh-open	|--x-|--x-|x-x-|--x-|
(4) hh-close	|x---|x---|----|x--x|
(5) cowbell	|x-x-|x-x-|x-x-|x-x-|
`,
		},
	}

	for _, exp := range tData {
		decoded, err := DecodeFile(path.Join("fixtures", exp.path))
		if err != nil {
			t.Fatalf("something went wrong decoding %s - %v", exp.path, err)
		}

		decoded.Version = "1.0-golang"
		decoded.Tempo = 240
		for j := 0; j < 16; j++ {
			decoded.Tracks[5].Set(j, j%2 == 0)
		}

		decoded.Write(path.Join("fixtures", exp.path+"_tmp"))
		defer os.Remove(path.Join("fixtures", exp.path+"_tmp"))

		modified, err := DecodeFile(path.Join("fixtures", exp.path+"_tmp"))
		if err != nil {
			t.Fatalf("something went wrong decoding %s - %v", exp.path, err)
		}

		if fmt.Sprint(modified) != exp.output {
			t.Logf("decoded:\n%#v\n", fmt.Sprint(modified))
			t.Logf("expected:\n%#v\n", exp.output)
			t.Fatalf("%s wasn't decoded as expect.\nGot:\n%s\nExpected:\n%s",
				exp.path, modified, exp.output)
		}
	}
}

func fileHash(path string, hasher hash.Hash) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := io.Copy(hasher, f); err != nil {
		return err
	}
	return nil
}

func fileSHA256(path string) string {
	hasher := sha256.New()
	fileHash(path, hasher)
	return hex.EncodeToString(hasher.Sum(nil))
}
