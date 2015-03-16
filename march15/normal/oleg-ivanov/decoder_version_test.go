package drum

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
)

func TestVersionShortcut(t *testing.T) {
	err := errors.New("Error")
	d := decoder{
		err: err,
	}
	r := io.Reader(strings.NewReader(""))
	d.version(&r)

	if d.err != err {
		t.Fatalf("Wrong error %+v for version shortcut", d.err)
	}
}

func TestVersionTooShort(t *testing.T) {
	d := decoder{}
	r := io.Reader(strings.NewReader("too short"))
	d.version(&r)

	if d.err == nil {
		t.Fatalf("Too short version failed")
	}
}

func TestVersionValid(t *testing.T) {
	b := [32]byte{}
	s := fmt.Sprintf("version%s", b[7:])

	d := decoder{}
	r := io.Reader(strings.NewReader(s))
	v := d.version(&r)

	if d.err != nil {
		t.Fatalf("Valid version errored with %+v", d.err)
	}
	if string(v) != "version" {
		t.Fatalf("Valid version incorrect: %s, expected %s", v, s)
	}
}
