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

func TestNonExistantFile(t *testing.T) {
	_, err := DecodeFile("nonexistant.splice")
	if !os.IsNotExist(err) {
		t.Fatalf("Expected ErrNotExist.  Got %v", err)
	}
}

func TestInvalidHeader(t *testing.T) {
	splice := []byte{
		0x53, 0x00, 0x4c, 0x49, 0x43, 0x45, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00,
	}
	_, err := DecodeByteArray(splice)
	if err.Error() != "Invalid Header" {
		t.Fatalf("Expected an invalid header error.  Got %v\n", err)
	}
}

func TestDecodeShortSplice(t *testing.T) {
	splice := []byte{
		0x53, 0x50, 0x4c, 0x49, 0x43, 0x45, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x3d, 0x30, 0x2e,
		0x39, 0x30, 0x39, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	_, err := DecodeByteArray(splice)
	if err.Error() != "EOF" {
		t.Fatalf("Expected EOF.  Got %v\n", err)
	}
}

func TestInvalidVersion(t *testing.T) {
	splice := []byte{
		0x53, 0x50, 0x4c, 0x49, 0x43, 0x45, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00,
	}
	_, err := DecodeByteArray(splice)
	if err.Error() != "Invalid Version" {
		t.Fatalf("Expected Invalid Version.  Got %v\n", err)
	}
}

func TestInvalidTrack(t *testing.T) {
	splice := []byte{
		0x53, 0x50, 0x4c, 0x49, 0x43, 0x45, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x3d, 0x30, 0x2e,
		0x39, 0x30, 0x39, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x78, 0x00, 0x00, 0x00, 0x00, 0x04, 'D',
		'R', 'U', 'M', 0x01, 0x02, 0x01, 0x01, 0x01,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00,
	}
	_, err := DecodeByteArray(splice)
	if err.Error() != "Invalid Track Format" {
		t.Fatalf("Expected Invalid Track.  Got %v\n", err)
	}
}
