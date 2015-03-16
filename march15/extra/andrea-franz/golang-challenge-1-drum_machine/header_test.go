package drum

import "testing"

func TestHeader_String(t *testing.T) {
	h := &Header{
		Filename:  "foo.splice",
		Signature: "SPLICE",
		Tempo:     200,
		Version:   "2 beta",
	}

	expected := "Saved with HW Version: 2 beta\nTempo: 200"
	s := h.String()
	if s != expected {
		t.Errorf("Header's String method returned an unexpected string.\nexpected:\n%s\ngot:\n%s", expected, s)
	}
}
