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

func TestDecodeFile_Empty(t *testing.T) {
	p, err := DecodeFile("")
	if err == nil || err.Error() != "open : no such file or directory" {
		t.Fatalf("should have file error.  Instead: %v", err)
	}
	if p != nil {
		t.Fatal("error, p should be nil")
	}

}
func TestDecode_Empty(t *testing.T) {
	p, err := Decode([]byte{})
	if err == nil || err.Error() != "Invalid file header" {
		t.Fatalf("should have invalid file header error.  Instead: %v", err)
	}
	if p != nil {
		t.Fatal("error, p should be nil")
	}
}

func TestDecode_MissingData(t *testing.T) {
	p, err := Decode(patternHeader)
	if err == nil || err.Error() != "Invalid file format, missing or invalid content length" {
		t.Fatalf("should have invalid file format error.  Instead: %v", err)
	}
	if p != nil {
		t.Fatal("error, p should be nil")
	}
}

func TestDecode_JunkData(t *testing.T) {
	d := append([]byte{}, patternHeader...)
	d = append(d, []byte("junk data blah blah blah")...)

	p, err := Decode(d)
	if err == nil || err.Error() != "Invalid file format, missing or invalid HWVersion" {
		t.Fatalf("should have invalid file format error.  Instead: %v", err)
	}
	if p != nil {
		t.Fatal("error, p should be nil")
	}
}

func TestDecode_BadTrack(t *testing.T) {
	d := append([]byte{}, patternHeader...)
	d = append(d, 0xF0)
	d = append(d, []byte("hwversion01234567890123456789012")...)
	d = append(d, 0x0, 0x0, 0x1, 0x0)

	p, err := Decode(d)
	if err == nil || err.Error() != "Invalid file format, missing or invalid track ID" {
		t.Fatalf("should have invalid file format error.  Instead: %v", err)
	}
	if p != nil {
		t.Fatal("error, p should be nil")
	}
}
