package drum

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io/ioutil"
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
		decoded, err := DecodeFile(path.Join("sequences", exp.path))
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

func TestBinaryMarshaling(t *testing.T) {
	paths := []string{
		"pattern_1.splice",
		"pattern_2.splice",
		"pattern_3.splice",
		"pattern_4.splice",
		"pattern_5.splice"}

	for _, fp := range paths {
		b, err := ioutil.ReadFile(path.Join("sequences", fp))
		if err != nil {
			t.Fatalf("something went wrong decoding %s - %v", fp, err)
		}
		p := &Pattern{}
		if err := p.UnmarshalBinary(b); err != nil {
			t.Fatalf("something went wrong decoding %s - %v", fp, err)
		}
		b2, err := p.MarshalBinary()
		if err != nil {
			t.Fatalf("something went wrong decoding %s - %v", fp, err)
		}
		if !bytes.Equal(b[:len(b2)], b2) {
			t.Logf("failed")
			fmt.Println(fp)
			fmt.Println(hex.Dump(b))
			fmt.Println("---")
			fmt.Println(hex.Dump(b2))
			fmt.Println()
			t.Fatalf("%s wasn't decoded as expect.", fp)
		}
	}
}
