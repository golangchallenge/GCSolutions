package drum

import (
	"bytes"
	"encoding/binary"
	"io"
	"os"
)

const (
	prefixln  = 0x0d
	versionln = 0x20
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
// TODO: implement
func DecodeFile(path string) (*Pattern, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return decode(f)
}

func decode(f *os.File) (*Pattern, error) {
	d := decoder{}
	p := Pattern{}

	b := d.data(f)
	p.Version = d.version(&b)
	p.Tempo = d.tempo(&b)

	for t := d.track(&b); t != nil; t = d.track(&b) {
		p.Tracks = append(p.Tracks, *t)
	}
	return &p, d.err
}

// decoder keeps track of the current state of decoding process,
// including global error state
type decoder struct {
	err error
}

// prefix describes structure of the prefix part, for easy
// reading by binary.Read
var prefix = struct {
	_ [prefixln]byte
}{}

// data returns main section of the pattern file, stripping away its prefix,
// and cutting data off at its declared length
func (d *decoder) data(f io.Reader) io.Reader {
	var b []byte

	if d.err = binary.Read(f, binary.LittleEndian, &prefix); d.err != nil {
		return nil
	}

	if b, d.err = pstring(f); d.err != nil {
		return nil
	}
	return bytes.NewReader(b)
}

// version of a hardware defined in the pattern file
func (d *decoder) version(r *io.Reader) []byte {
	if d.err != nil {
		return nil
	}

	b := make([]byte, versionln)
	if d.err = binary.Read(*r, binary.LittleEndian, b); d.err != nil {
		return nil
	}

	return bytes.TrimRight(b, "\x00")
}

// tempo defined in the pattern file
func (d *decoder) tempo(r *io.Reader) Tempo {
	var t float32

	if d.err != nil {
		return Tempo(t)
	}

	d.err = binary.Read(*r, binary.LittleEndian, &t)
	return Tempo(t)
}

// track parsing
func (d *decoder) track(r *io.Reader) *Track {
	if d.err != nil {
		return nil
	}

	t := Track{}
	err := binary.Read(*r, binary.LittleEndian, &t.ID)
	// if decoder reached end of data - it will be indicated by io.EOF
	// error value, but it should not be reported as a global decoder error.
	// Other values are global errors, though
	if err != io.EOF {
		d.err = err
	}
	if err != nil {
		return nil
	}

	if t.Name, d.err = pstring(*r); d.err != nil {
		return nil
	}

	if d.err = binary.Read(*r, binary.LittleEndian, &t.Steps); d.err != nil {
		return nil
	}
	return &t
}
