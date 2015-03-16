package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestTempoShortcut(t *testing.T) {
	err := errors.New("Error")
	d := decoder{
		err: err,
	}
	r := io.Reader(strings.NewReader(""))
	d.tempo(&r)

	if d.err != err {
		t.Fatalf("Wrong error %+v for tempo shortcut", d.err)
	}
}

func TestTempoTooShort(t *testing.T) {
	d := decoder{}
	r := io.Reader(strings.NewReader("xx"))
	d.tempo(&r)

	if d.err == nil {
		t.Fatalf("Too short tempo failed")
	}
}

func TestTempoValid(t *testing.T) {
	f := float32(100.23)
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.LittleEndian, f)

	d := decoder{}
	r := io.Reader(strings.NewReader(buf.String()))
	m := d.tempo(&r)

	if d.err != nil {
		t.Fatalf("Valid tempo errored with %+v", d.err)
	}
	if m != Tempo(f) {
		t.Fatalf("Valid tempo incorrect: %+v", m)
	}
}
