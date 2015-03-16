package drum

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"testing"
)

func TestUpdatePattern(t *testing.T) {
	filename := "pattern_2.splice"

	decoded, err := DecodeFile(path.Join("fixtures", filename))
	if err != nil {
		t.Fatalf("something went wrong decoding %s - %v", filename, err)
	}

	// get pointer to the cowbell track
	cowbell := &decoded.Tracks[3].Steps

	// add more cowbell
	cowbell[0] = 1
	cowbell[4] = 1
	cowbell[6] = 1
	cowbell[12] = 1
	cowbell[14] = 1

	expected := `Saved with HW Version: 0.808-alpha
Tempo: 98.4
(0) kick	|x---|----|x---|----|
(1) snare	|----|x---|----|x---|
(3) hh-open	|--x-|--x-|x-x-|--x-|
(5) cowbell	|x---|x-x-|x---|x-x-|
`

	if fmt.Sprint(decoded) != expected {
		t.Logf("decoded:\n%#v\n", fmt.Sprint(decoded))
		t.Logf("expected:\n%#v\n", expected)
		t.Fatalf("%s wasn't decoded as expect.\nGot:\n%s\nExpected:\n%s", filename, decoded, expected)
	}
}

func TestEncodeWriter(t *testing.T) {
	filename := "pattern_2.splice"

	decoded, err := DecodeFile(path.Join("fixtures", filename))
	if err != nil {
		t.Fatalf("something went wrong decoding %s - %v", filename, err)
	}

	cowbell := &decoded.Tracks[3].Steps

	// add more cowbell
	cowbell[0] = 1
	cowbell[4] = 1
	cowbell[6] = 1
	cowbell[12] = 1
	cowbell[14] = 1

	// binary encode
	var buffer bytes.Buffer

	_, err = EncodePattern(decoded, &buffer)
	if err != nil {
		t.Fatalf("something went wrong encoding %s - %v", filename, err)
	}

	// decode again
	decoded2, err := DecodePattern(&buffer)
	if err != nil {
		t.Fatalf("something went wrong encoding %s - %v", filename, err)
	}

	expected := `Saved with HW Version: 0.808-alpha
Tempo: 98.4
(0) kick	|x---|----|x---|----|
(1) snare	|----|x---|----|x---|
(3) hh-open	|--x-|--x-|x-x-|--x-|
(5) cowbell	|x---|x-x-|x---|x-x-|
`

	if fmt.Sprint(decoded2) != expected {
		t.Logf("decoded:\n%#v\n", fmt.Sprint(decoded))
		t.Logf("expected:\n%#v\n", expected)
		t.Fatalf("%s wasn't decoded as expect.\nGot:\n%s\nExpected:\n%s", filename, decoded, expected)
	}
}

func TestEncodeFile(t *testing.T) {
	pattern := Pattern{
		"0.1210-alpha",
		57,
		[]Track{
			{0, "kick", Measure{
				1, 0, 0, 0,
				0, 0, 0, 0,
				1, 0, 0, 0,
				0, 0, 0, 0,
			}},
			{1, "snare", Measure{
				0, 0, 0, 0,
				1, 0, 0, 0,
				0, 0, 0, 0,
				1, 0, 0, 0,
			}},
			{3, "hh-open", Measure{
				0, 0, 1, 0,
				0, 0, 1, 0,
				1, 0, 1, 0,
				0, 0, 1, 0,
			}},
			{5, "cowbell", Measure{
				1, 0, 0, 0,
				1, 0, 1, 0,
				1, 0, 0, 0,
				1, 0, 1, 0,
			}},
		},
	}

	filename := "test_pattern.splice"

	defer os.Remove(filename)
	if err := EncodeFile(&pattern, filename); err != nil {
		t.Fatalf("something went wrong encoding %s - %v", filename, err)
	}

	decoded, err := DecodeFile(filename)
	if err != nil {
		t.Fatalf("something went wrong decoding %s - %v", filename, err)
	}

	expected := `Saved with HW Version: 0.1210-alpha
Tempo: 57
(0) kick	|x---|----|x---|----|
(1) snare	|----|x---|----|x---|
(3) hh-open	|--x-|--x-|x-x-|--x-|
(5) cowbell	|x---|x-x-|x---|x-x-|
`

	if fmt.Sprint(decoded) != expected {
		t.Logf("decoded:\n%#v\n", fmt.Sprint(decoded))
		t.Logf("expected:\n%#v\n", expected)
		t.Fatalf("%s wasn't decoded as expect.\nGot:\n%s\nExpected:\n%s", filename, decoded, expected)
	}
}
