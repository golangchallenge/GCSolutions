package drum

import (
	"strings"
	"testing"
)

func TestPstringErrors(t *testing.T) {
	xs := []string{
		"",
		"\x05",
		"\x10Boom",
	}
	for _, x := range xs {
		r := strings.NewReader(x)
		if _, err := pstring(r); err == nil {
			t.Errorf("Should have errored on empty input")
		}
	}
}

func TestPstringValid(t *testing.T) {
	var s []byte
	var err error

	xs := map[string]string{
		"\x04Boom": "Boom",
		"\x00":     "",
		"\x03Boom": "Boo",
	}

	for k, v := range xs {
		r := strings.NewReader(k)
		if s, err = pstring(r); err != nil {
			t.Errorf("Got %q, expected %q for %q", s, v, r)
		} else if v != string(s) {
			t.Errorf("Got %q, expected %q for %q", s, v, r)
		}
	}

}
