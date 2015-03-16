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


func TestTrackToString(t *testing.T) {
  track := &track{
    ID:   10,
    Name: "cool-sound",
    Steps: []byte{
      1, 1, 0, 0,
      0, 0, 1, 1,
      1, 1, 0, 0,
      0, 0, 1, 1,
    },
  }

  expected := fmt.Sprintf("(10) cool-sound\t|xx--|--xx|xx--|--xx|\n")
  if track.String() != expected {
    t.Logf("Expected: %#v", expected)
    t.Logf("Got:      %#v", track.String())
    t.Fatalf("Track wasn't represented as expected.\nGot:\n%s\nExpected:\n%s",
      track.String(), expected)
  }
}

func TestHeaderToString(t *testing.T) {
  header := &patternHeader{
    Version: [16]byte{
      97, 98, 99, 100, 0, 0, 0, 0,
      0, 0, 0, 0, 0, 0, 0, 0,
    },
    Tempo: 240,
  }

  expected := fmt.Sprintf("Saved with HW Version: abcd\nTempo: 240\n")
  if header.String() != expected {
    t.Logf("Expected: %#v", expected)
    t.Logf("Got:      %#v", header.String())
    t.Fatalf("Header wasn't represented as expected.\nGot:\n%s\nExpected:\n%s",
      header.String(), expected)
  }
}

