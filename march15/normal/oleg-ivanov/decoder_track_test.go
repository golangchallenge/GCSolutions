package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
)

func TestTrackShortcut(t *testing.T) {
	err := errors.New("Error")
	d := decoder{
		err: err,
	}
	r := io.Reader(strings.NewReader(""))
	d.track(&r)

	if d.err != err {
		t.Fatalf("Wrong error %+v for track shortcut", d.err)
	}
}

func TestTrackAtEOF(t *testing.T) {
	d := decoder{}
	r := io.Reader(strings.NewReader(""))
	k := d.track(&r)

	if d.err != nil {
		t.Fatalf("Track at EOF failed: %+v", d.err)
	}
	if k != nil {
		t.Fatalf("Track at EOF returned non-nil: %+v", k)
	}

}

func TestTrackNoID(t *testing.T) {
	d := decoder{}
	r := io.Reader(strings.NewReader("xx"))
	d.track(&r)

	if d.err == nil {
		t.Fatalf("Track with no ID failed")
	}
}

func TestTrackNameCutoff(t *testing.T) {
	id := int32(11)
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.LittleEndian, id)

	s := fmt.Sprintf("%s%s", buf.String(), "\x55short")

	d := decoder{}
	r := io.Reader(strings.NewReader(s))
	d.track(&r)

	if d.err == nil {
		t.Fatalf("Track with name cut-off failed")
	}
}

func TestTrackStepsCutoff(t *testing.T) {
	id := int32(11)
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.LittleEndian, id)

	s := fmt.Sprintf("%s%s", buf.String(), "\x05short\x00\x00")

	d := decoder{}
	r := io.Reader(strings.NewReader(s))
	d.track(&r)

	if d.err == nil {
		t.Fatalf("Track with steps cut-off failed")
	}
}

func TestTrackValid(t *testing.T) {
	steps := [16]byte{
		0, 0, 0, 1,
		0, 0, 0, 1,
		0, 0, 0, 1,
		0, 0, 0, 1,
	}
	id := int32(11)
	name := "drum"
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.LittleEndian, id)

	s := fmt.Sprintf("%s%c%s%s", buf.String(), len(name), name, steps)

	d := decoder{}
	r := io.Reader(strings.NewReader(s))
	k := d.track(&r)

	if d.err != nil {
		t.Fatalf("Valid track errored: %+v", d.err)
	}

	if k == nil {
		t.Fatalf("Valid track returned as nil track")
	}

	if k.ID != id {
		t.Fatalf("Valid track ID mismatch: %d, expected %d", k.ID, id)
	}

	if string(k.Name) != "drum" {
		t.Fatalf("Valid track name mismatch: %d, expected %d", k.ID, id)
	}

	if k.Steps != Steps(steps) {
		t.Fatalf("Valid track steps mismatch: %q, expected %q",
			k.Steps, steps)
	}
}
