package drum

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
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

func TestDecodeFileErrors(t *testing.T) {
	tests := []struct {
		n   int64
		err error
	}{
		// In these tests, we limit the reads to n bytes. This should give
		// us an error to test our error cases.
		{n: 1, err: ErrNotSpliceFile},     // 0 - header.
		{n: 13, err: ErrNotSpliceFile},    // 1 - size.
		{n: 25, err: ErrNotSpliceFile},    // 2 - version.
		{n: 48, err: ErrNotSpliceFile},    // 3 - tempo.
		{n: 53, err: io.ErrUnexpectedEOF}, // 4 - track id.
		{n: 54, err: io.EOF},              // 5 - track name length.
		{n: 57, err: io.ErrUnexpectedEOF}, // 6 - track name.
		{n: 60, err: io.EOF},              // 7 - track steps.
	}

	data, err := ioutil.ReadFile(path.Join("fixtures", "pattern_1.splice"))
	if err != nil {
		t.Fatalf("failed to read fixtures/pattern_1.plice: %v", err)
	}

	for k, test := range tests {
		buf := bytes.NewBuffer(data)
		lr := &io.LimitedReader{
			R: buf,
			N: test.n,
		}
		_, err := decodePattern(lr)
		if err != test.err {
			t.Errorf("Test %v: unexpected error '%v': expected %v",
				k, err, test.err)
		}
	}

	// Test the inability to open a file.
	if _, err := DecodeFile("not-exist.txt"); !os.IsNotExist(err) {
		t.Errorf("Didn't get ErrNotExist when trying to decode a file "+
			"that doesn't exist: %v", err)
	}

	// Test a file that doesn't start with SPLICE.
	if _, err := DecodeFile("decoder.go"); err != ErrNotSpliceFile {
		t.Errorf("Didn't get ErrNotSpliceFile when trying to decode a "+
			"non splice file: %v", err)
	}
}

func TestEncodePattern(t *testing.T) {
	tests := []string{
		"pattern_1.splice",
		"pattern_2.splice",
		"pattern_3.splice",
		"pattern_4.splice",
	}

	for _, test := range tests {
		// Get the original file.
		org, err := ioutil.ReadFile(path.Join("fixtures", test))
		if err != nil {
			t.Fatalf("failed to open %v: %v", test, err)
		}

		// Generate a pattern from the file.
		p, err := decodePattern(bytes.NewBuffer(org))
		if err != nil {
			t.Fatalf("failed to decode %v to a pattern: %v", test, err)
		}

		// Write out the test file.
		f, err := ioutil.TempFile("", "TestEncodePattern")
		if err != nil {
			t.Fatalf("failed to make test file.")
		}
		f.Close()
		defer os.Remove(f.Name())
		if err := EncodeFile(p, f.Name()); err != nil {
			t.Fatalf("failed to write '%v' to test file.", test)
		}

		// Get the data from the test file.
		buf, err := ioutil.ReadFile(f.Name())
		if err != nil {
			t.Fatalf("failed to open %v: %v", f.Name(), err)
		}

		// Ensure the original and the test file are the same.
		if bytes.Compare(org, buf) != 0 {
			t.Errorf("encoded value does not equal original.")
			t.Logf("original:\n")
			t.Logf(hex.Dump(org))
			t.Logf("result:\n")
			t.Logf(hex.Dump(buf))

		}
	}

}

func TestEncodePatternErrors(t *testing.T) {
	// Test a file that doesn't exists.
	if err := EncodeFile(&Pattern{}, "/something/that/should/fail/"); err == nil {
		t.Errorf("EncodeFile didn't faile with a bad filename.")
	}

	// Test a write error.
	ew := &errWriter{
		w:   &bytes.Buffer{},
		err: io.EOF,
	}
	if n, err := ew.Write([]byte("test")); n != 0 || err != io.EOF {
		t.Errorf("The errWriter didn't return an error when err was set.")
	}
}
