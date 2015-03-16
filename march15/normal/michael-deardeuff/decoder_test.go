package drum

import (
	"fmt"
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
		{"pattern_6.splice",
			`Saved with HW Version: 0.808-alpha
Tempo: 120
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

func TestDecodeFile_goodFiles(t *testing.T) {
	tData := []string{
		"pattern_5.splice",
		"pattern_6.splice",
		"pattern_7.splice",
		"pattern_8_paul_is_dead.splice",
	}
	for _, name := range tData {
		_, err := DecodeFile(path.Join("fixtures", name))
		if err != nil {
			t.Errorf("something went wrong decoding %s - %v", name, err)
		}
	}
}

func TestDecodeFile_badFiles(t *testing.T) {
	tData := []string{
		"this-file-does-not-exist.splice",

		"bad_len_too_short.splice",
		"bad_len_too_long.splice",
		"bad_not_splice.splice",
		"bad_truncated_0.splice",
		"bad_truncated_1.splice",
		"bad_truncated_2.splice",
		"bad_truncated_3.splice",
		"bad_truncated_4.splice",
		"bad_truncated_5.splice",
		"bad_truncated_6.splice",
		"bad_truncated_7.splice",
		"bad_truncated_8.splice",
		"bad_truncated_9.splice",
		"bad_truncated_A.splice",
		"bad_truncated_B.splice",
		"bad_truncated_C.splice",
		"bad_truncated_D.splice",
	}
	for i, name := range tData {
		fullPath := path.Join("fixtures", name)

		if _, err := os.Stat(fullPath); i != 0 && os.IsNotExist(err) {
			t.Errorf("Missing test file %s", name)
			continue
		}

		decoded, err := DecodeFile(fullPath)
		if err == nil {
			t.Errorf("nothing went wrong decoding! %s - %v", name, decoded)
		} else if decoded != nil {
			t.Errorf("bad encoding still returned a pattern! %s - %v", name, decoded)
		}
	}
}
