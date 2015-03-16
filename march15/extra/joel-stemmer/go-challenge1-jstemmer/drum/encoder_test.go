package drum

import (
	"bytes"
	"encoding/hex"
	"io/ioutil"
	"path"
	"testing"
)

func TestEncode(t *testing.T) {
	tests := []struct {
		path string
	}{
		{"pattern_1.splice"},
		{"pattern_2.splice"},
		{"pattern_3.splice"},
		{"pattern_4.splice"},
		{"pattern_5.splice"},
	}

	for _, tt := range tests {
		data, err := ioutil.ReadFile(path.Join("fixtures", tt.path))
		if err != nil {
			t.Errorf("could not read %s: %v", tt.path, err)
			continue
		}

		var p Pattern
		err = NewDecoder(bytes.NewReader(data)).Decode(&p)
		if err != nil {
			t.Errorf("could not decode %s: %v", tt.path, err)
			continue
		}

		var buf bytes.Buffer
		err = NewEncoder(&buf).Encode(p)
		if err != nil {
			t.Errorf("could not encode %s: %v", tt.path, err)
			continue
		}

		if !bytes.Equal(data, buf.Bytes()) {
			t.Errorf("incorrect encoded pattern.\nGot:\n%s\nWant:\n%s\n", hex.Dump(buf.Bytes()), hex.Dump(data))
		}
	}
}
