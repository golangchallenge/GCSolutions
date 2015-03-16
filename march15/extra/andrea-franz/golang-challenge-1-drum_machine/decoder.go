package drum

import (
	"io"
	"os"
	"path"
)

// A Decoder reads and decodes Patterns from an input stream.
type Decoder struct {
	r io.Reader
}

// Decode reads the next Pattern from its input and stores it in the value pointed to by p.
func (d *Decoder) Decode(p *Pattern) error {
	s := newScanner(d.r, p)
	if p.Header == nil {
		p.Header = &Header{}
	}

	return s.unmarshal(p)
}

// NewDecoder returns a new decoder that reads from r.
func NewDecoder(r io.Reader) *Decoder {
	d := &Decoder{
		r: r,
	}

	return d
}

// DecodeFile decodes a file and return a Pattern.
func DecodeFile(filepath string) (*Pattern, error) {
	p := NewPattern()
	p.Header.Filename = path.Base(filepath)
	f, err := os.Open(filepath)
	if err != nil {
		return p, err
	}
	defer f.Close()

	err = NewDecoder(f).Decode(p)
	return p, err
}
