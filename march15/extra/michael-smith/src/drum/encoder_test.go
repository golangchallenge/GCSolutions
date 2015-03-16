package drum

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestEncodeFile(t *testing.T) {
	files := []string{
		"pattern_1.splice",
		"pattern_2.splice",
		"pattern_3.splice",
		"pattern_4.splice",
		// "pattern_5.splice", // skipped because it has corruption at the end
	}
	for _, file := range files {
		f, err := os.Open(path.Join("fixtures", file))
		if err != nil {
			t.Fatalf("something went wrong reading file %s - %v", file, err)
		}
		defer f.Close()

		var fileBytes []byte
		if fileBytes, err = ioutil.ReadAll(f); err != nil {
			t.Fatalf("something went wrong reading file %s - %v", file, err)
		}

		pattern, err := Decode(bytes.NewReader(fileBytes))
		if err != nil {
			t.Fatalf("something went wrong decoding %s - %v", file, err)
		}

		buf := new(bytes.Buffer)
		if err := Encode(pattern, buf); err != nil {
			t.Fatalf("something went wrong encoding %s - %v", file, err)
		}
		encodedBytes := buf.Bytes()

		if !bytes.Equal(fileBytes, encodedBytes) {
			t.Logf("expected:\n% x\n", fileBytes)
			t.Logf("got:\n% x\n", encodedBytes)
			t.Fatalf("%s doesn't match expected bytes\n", file)
		}
	}
}
