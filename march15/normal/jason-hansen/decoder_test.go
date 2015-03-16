package drum

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"reflect"
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

func TestDecodeEmptyFile(t *testing.T) {
	if _, err := DecodeFile(path.Join("fixtures", "empty.splice")); err == nil {
		t.Fatalf("no error returned while trying to decode an empty file")
	}
}

func TestDecodeNoSuchFile(t *testing.T) {
	_, err := DecodeFile("filethatdoesnotexist")
	if err == nil {
		t.Fatalf("no error returned while trying to decode a file that doesn't exist")
	}

	if _, ok := err.(*os.PathError); !ok {
		t.Fatalf("wrong error returned while trying to decode a file that doesn't exist"+
			"\nGot:\n%s\nExpected:\n%s", reflect.TypeOf(err), reflect.TypeOf(&os.PathError{}))
	}
}

func TestDecode(t *testing.T) {
	tData := []*struct {
		input  []byte
		output Pattern
	}{
		{[]byte{
			// Signature
			'S', 'P', 'L', 'I', 'C', 'E',
			// Remaining length
			0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3d,
			// Version
			'0', '.', '8', '0', '9', 0x0, 0x0, 0x0,
			0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
			0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
			0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
			// Tempo
			0x0, 0x0, 0xf0, 0x42,
			// Track ID
			0x0,
			// Track name length
			0x0, 0x0, 0x0, 0x04,
			// Track name
			'k', 'i', 'c', 'k',
			// Track steps
			0x1, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0,
			0x1, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0,
		}, Pattern{version: "0.809", tempo: float32(120),
			tracks: []track{{
				id: 0, name: "kick",
				steps: [16]step{
					1, 0, 0, 0,
					1, 0, 0, 0,
					1, 0, 0, 0,
					1, 0, 0, 0,
				},
			},
			}},
		},
		{[]byte{
			// Signature
			'S', 'P', 'L', 'I', 'C', 'E',
			// Remaining length
			0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x5a,
			// Version
			'0', '.', '8', '0', '8', '-', 'a', 'l',
			'p', 'h', 'a', 0x0, 0x0, 0x0, 0x0, 0x0,
			0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
			0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
			// Tempo
			0xcd, 0xcc, 0xc4, 0x42,
			// Track ID
			0x3,
			// Track name length
			0x0, 0x0, 0x0, 0x05,
			// Track name
			's', 'n', 'a', 'r', 'e',
			// Track steps
			0x1, 0x0, 0x1, 0x0, 0x1, 0x0, 0x1, 0x0,
			0x1, 0x0, 0x1, 0x0, 0x1, 0x0, 0x1, 0x0,
			// Track ID
			0x8,
			// Track name length
			0x0, 0x0, 0x0, 0x07,
			// Track name
			'c', 'o', 'w', 'b', 'e', 'l', 'l',
			// Track steps
			0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1,
			0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		}, Pattern{version: "0.808-alpha", tempo: float32(98.4),
			tracks: []track{
				{
					id: 3, name: "snare",
					steps: [16]step{
						1, 0, 1, 0,
						1, 0, 1, 0,
						1, 0, 1, 0,
						1, 0, 1, 0,
					},
				},
				{
					id: 8, name: "cowbell",
					steps: [16]step{
						1, 1, 1, 1,
						1, 1, 1, 1,
						0, 0, 0, 0,
						0, 0, 0, 0,
					},
				},
			}},
		},
	}

	for _, d := range tData {
		r := bytes.NewReader(d.input)
		dec := NewPatternDecoder(r)
		var pat Pattern
		err := dec.Decode(&pat)
		if err != nil {
			t.Fatalf("something went wrong decoding pattern data - %v", err)
		}

		if !reflect.DeepEqual(&pat, &d.output) {
			t.Fatalf("pattern was decoded incorrectly"+
				"\nGot:\n%s\nExpected:\n%s", &pat, &d.output)
		}
	}
}

func TestDecodeBadSignature(t *testing.T) {
	tData := []byte{'B', 'A', 'D', 'S', 'I', 'G'}
	r := bytes.NewReader(tData)
	dec := NewPatternDecoder(r)
	var pat Pattern
	err := dec.Decode(&pat)

	if err == nil {
		t.Fatalf("no error was returned after decoding a bad signature")
	}
	if err != ErrInvalidSignature {
		t.Fatalf("wrong error was returned after decoding a bad signature"+
			"\nGot:\n%v\nExpected:\n%v", err, ErrInvalidSignature)
	}
}

func TestDecodeBadStep(t *testing.T) {
	tData := []byte{
		// Signature
		'S', 'P', 'L', 'I', 'C', 'E',
		// Remaining length
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3d,
		// Version
		'0', '.', '8', '0', '9', 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		// Tempo
		0x0, 0x0, 0xf0, 0x42,
		// Track ID
		0x0,
		// Track name length
		0x0, 0x0, 0x0, 0x04,
		// Track name
		'k', 'i', 'c', 'k',
		// Track steps (one of the steps is 5, which is invalid)
		0x1, 0x0, 0x0, 0x5, 0x1, 0x0, 0x0, 0x0,
		0x1, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0,
	}
	r := bytes.NewReader(tData)
	dec := NewPatternDecoder(r)
	var pat Pattern
	err := dec.Decode(&pat)

	if err == nil {
		t.Fatalf("no error was returned after decoding a bad signature")
	}
	if err != ErrInvalidStep {
		t.Fatalf("wrong error was returned after decoding a bad signature"+
			"\nGot:\n%v\nExpected:\n%v", err, ErrInvalidStep)
	}
}

func TestUnexpectedEOF(t *testing.T) {
	tData := []byte{
		// Signature
		'S', 'P', 'L', 'I', 'C', 'E',
		// Remaining length
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3d,
		// Version
		'0', '.', '8', '0', '9', 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		// Tempo
		0x0, 0x0, 0xf0, 0x42,
		// Track ID
		0x0,
		// Track name length
		0x0, 0x0, 0x0, 0x04,
		// Track name
		'k', 'i', 'c', 'k',
		// Track steps (one of the steps is 5, which is invalid)
		0x1, 0x0, 0x0, 0x1, 0x1, 0x0, 0x0, 0x0,
		// Everything is valid up to this point,
		// but now some of the steps are missing here.
	}
	r := bytes.NewReader(tData)
	dec := NewPatternDecoder(r)
	var pat Pattern
	err := dec.Decode(&pat)

	if err == nil {
		t.Fatalf("no error was returned after decoding a bad signature")
	}
	if err != ErrUnexpectedEOF {
		t.Fatalf("wrong error was returned after decoding a bad signature"+
			"\nGot:\n%v\nExpected:\n%v", err, ErrUnexpectedEOF)
	}
}

func TestPrivateDecode(t *testing.T) {
	dec := &PatternDecoder{r: bytes.NewReader([]byte("TEST")), remaining: 10}

	var tBytes [4]byte
	err := dec.decode(&tBytes, nil)
	if err != nil {
		t.Fatalf("something went wrong decoding byte array - %v", err)
	}

	const expRem = 6
	if dec.remaining != expRem {
		t.Fatalf("remaining on PatternDecoder not updated correctly"+
			"\nGot:\n%v\nExpected:\n%v", dec.remaining, expRem)
	}
}

func TestDecodeRemaining(t *testing.T) {
	dec := &PatternDecoder{r: bytes.NewReader([]byte("TEST")), remaining: 10}

	var tBytes [4]byte
	err := dec.decodeIgnoreRemaining(&tBytes, nil)
	if err != nil {
		t.Fatalf("something went wrong decoding byte array - %v", err)
	}

	const expRem = 10
	if dec.remaining != expRem {
		t.Fatalf("remaining on PatternDecoder changed after call to decodeIgnoreRemaining"+
			"\nGot:\n%v\nExpected:\n%v", dec.remaining, expRem)
	}
}
