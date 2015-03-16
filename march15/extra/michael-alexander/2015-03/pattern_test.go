package drum

import (
	"bytes"
	"encoding/hex"
	"io/ioutil"
	"path"
	"testing"
)

func TestPattern_Encode(t *testing.T) {
	for _, f := range []string{
		"pattern_1.splice",
		"pattern_2.splice",
		"pattern_3.splice",
		"pattern_4.splice",
		"pattern_5.splice",
	} {
		raw, err := ioutil.ReadFile(path.Join("fixtures", f))
		if err != nil {
			t.Fatalf("unable to read %s, %v", f, err)
		}

		decoded, err := Decode(bytes.NewBuffer(raw))
		if err != nil {
			t.Fatalf("unable to decode %s, %v", f, err)
		}

		encoded := bytes.NewBuffer([]byte{})
		if err := decoded.Encode(encoded); err != nil {
			t.Fatalf("unable to encode %s, %v", f, err)
		}

		rawEncoded := encoded.Bytes()
		if !bytes.HasPrefix(raw, rawEncoded) {
			t.Errorf(
				"encoded did not match raw for %s.\nExpected:\n\n%s\n\nActual:\n\n%s",
				f,
				hex.Dump(raw),
				hex.Dump(rawEncoded),
			)
		}
	}
}
