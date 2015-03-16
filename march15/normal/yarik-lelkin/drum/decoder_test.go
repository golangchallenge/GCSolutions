package drum

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path"
	"testing"
)

var tData = []struct {
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

func TestEncodeFile(t *testing.T) {
	data, err := ioutil.ReadFile(path.Join("fixtures", tData[0].path))
	if err != nil {
		t.Fatalf("cannot open file %s", tData[0].path)
	}
	p, err := Decode(data)
	if err != nil {
		t.Fatalf("something went wrong decoding %s - %v", tData[0].path, err)
	}

	data1, err := p.MarshalBinary()
	if err != nil {
		t.Fatalf("something went wrong encoding %s - %v", tData[0].path, err)
	}
	if bytes.Compare(data, data1) != 0 {
		t.Fatalf("original and marshalled binaries for %s does not match", tData[0].path)
	}
}

func TestDecodePartiallyCorryptedFile(t *testing.T) {
	data0, err := ioutil.ReadFile(path.Join("fixtures", tData[0].path))
	if err != nil {
		t.Fatalf("cannot open file %s", tData[0].path)
	}
	data1, err := ioutil.ReadFile(path.Join("fixtures", tData[1].path))
	if err != nil {
		t.Fatalf("cannot open file %s", tData[1].path)
	}
	data2, err := ioutil.ReadFile(path.Join("fixtures", tData[2].path))
	if err != nil {
		t.Fatalf("cannot open file %s", tData[2].path)
	}

	dataCorrupted := data0[:len(data0)/2]
	dataCorrupted = append(dataCorrupted, data1...)
	dataCorrupted = append(dataCorrupted, data2[:len(data2)/2]...)

	var p *Pattern
	if p, err = Decode(dataCorrupted); err != nil {
		t.Fatalf("something went wrong decoding partially corrupted %s - %v", tData[0].path, err)
	}

	if fmt.Sprint(p) != tData[1].output {
		t.Logf("decoded:\n%#v\n", p)
		t.Logf("expected:\n%#v\n", tData[1].output)
		t.Fatalf("Partially corrupted %s wasn't decoded as expect.\nGot:\n%s\nExpected:\n%s",
			tData[1].path, p, tData[1].output)
	}
}

func TestDecodeFile(t *testing.T) {

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
