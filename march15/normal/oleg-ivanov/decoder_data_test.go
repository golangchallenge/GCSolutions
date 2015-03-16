package drum

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func testData(s string) decoder {
	d := decoder{}
	d.data(strings.NewReader(s))
	return d
}

func TestDataShorterThanPrefix(t *testing.T) {
	d := testData("too short")
	if d.err == nil {
		t.Fatalf("Data file shorter than prefix failed")
	}
}

func TestDataPrefixOnly(t *testing.T) {
	pfx := [0x0d]byte{}
	s := fmt.Sprintf("%s", pfx)

	if testData(s).err == nil {
		t.Fatalf("Prefix-only data file failed")
	}
}

func TestDataShortContent(t *testing.T) {
	pfx := [0x0d]byte{}
	s := fmt.Sprintf("%s\xc0content", pfx)

	if testData(s).err == nil {
		t.Fatalf("Short-content data file failed")
	}
}

func TestDataValid(t *testing.T) {
	pfx := [0x0d]byte{}
	data := "some content"
	s := fmt.Sprintf("%s%c%s", pfx, len(data), data)

	d := decoder{}
	r := d.data(strings.NewReader(s))

	if d.err != nil {
		t.Fatalf("Valid data errored")
	}

	b := new(bytes.Buffer)
	b.ReadFrom(r)
	if b.String() != data {
		t.Fatalf("Valid data failed to read")
	}
}
