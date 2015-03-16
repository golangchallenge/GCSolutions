package drum

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func fixture(name string) []byte {
	path := fmt.Sprintf("fixtures/%s.splice", name)
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}

	return b
}

func TestEncoder(t *testing.T) {
	expected := fixture("pattern_1")
	p, err := DecodeFile("fixtures/pattern_1.splice")
	if err != nil {
		log.Fatal(err)
	}

	out := bytes.NewBuffer([]byte{})
	err = NewEncoder(out).Encode(p)
	if err != nil {
		log.Fatal(err)
	}

	data := out.Bytes()
	if bytes.Compare(expected, data) != 0 {
		t.Fatalf("Encoding error.\nExpected:\n%v\nGot:\n%v\n", expected, data)
	}
}

func TestEncoder_Changes(t *testing.T) {
	p, err := DecodeFile("fixtures/pattern_2.splice")
	if err != nil {
		log.Fatal(err)
	}

	p.Header.Filename = "pattern_2-morebells.splice"
	cowbell := p.Tracks[3]
	cowbell.SetStep(0, true)
	cowbell.SetStep(4, true)
	cowbell.SetStep(6, true)
	cowbell.SetStep(12, true)
	cowbell.SetStep(14, true)

	out := bytes.NewBuffer([]byte{})
	err = NewEncoder(out).Encode(p)
	if err != nil {
		log.Fatal(err)
	}

	p2 := NewPattern()
	err = NewDecoder(out).Decode(p2)
	if err != nil {
		log.Fatal(err)
	}

	expected := `Saved with HW Version: 0.808-alpha
Tempo: 98.4
(0) kick	|x---|----|x---|----|
(1) snare	|----|x---|----|x---|
(3) hh-open	|--x-|--x-|x-x-|--x-|
(5) cowbell	|x---|x-x-|x---|x-x-|
`

	if fmt.Sprint(p2) != expected {
		t.Logf("decoded:\n%#v\n", fmt.Sprint(p2))
		t.Logf("expected:\n%#v\n", expected)
		t.Fatalf("pattern wasn't changed as expect.\nGot:\n%s\nExpected:\n%s", p2, expected)
	}
}
