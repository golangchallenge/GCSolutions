// Package drum is supposed to implement the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
package drum

import (
	"bytes"
	"encoding/binary"
	"io"

	"os"
	"text/template"
)

var (
	pt *template.Template
)

const (
	pDoc = `Saved with HW Version: {{.Version}}
Tempo: {{.Tempo}}
{{ range .Instruments }}{{ . }}
{{ end }}`
)

func init() {
	pt = template.Must(template.New("pattern").Parse(pDoc))
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	PatternHeader
	Instruments []Instrument
	fp          *os.File
}

func (p *Pattern) String() string {
	var buf bytes.Buffer
	pt.Execute(&buf, p)
	return buf.String()
}

// PatternHeader represents the fixed-width data
// for a Pattern (can be loaded with binary.Read).
type PatternHeader struct {
	Splice  [13]byte
	Length  uint8
	Version Version
	Tempo   float32
}

// Version represents the fixed-width version data
// from .splice files.
type Version [32]byte

func (v *Version) String() string {
	return string(bytes.Trim(v[:], "\x00"))
}

// NewPattern Takes a file pointer  and creates a
// Pattern.  The pattern gets parsed with its
// parser functions (see 'type patternParser' above).
func NewPattern(fp *os.File) (*Pattern, error) {
	p := &Pattern{fp: fp}
	if err := binary.Read(p.fp, binary.LittleEndian, &p.PatternHeader); err != nil {
		return p, err
	}
	return p, parseInstruments(p)
}

func parseInstruments(p *Pattern) error {
	for keepGoing(p) {
		i, err := newInstrument(p.fp)
		if err != nil {
			return err
		}
		p.Instruments = append(p.Instruments, i)
	}
	return nil
}

// Checks if the file pointer has more data from
// the .splice file
func keepGoing(p *Pattern) bool {
	pos, err := p.fp.Seek(0, 1)
	return pos < int64(p.Length) && err != io.EOF
}
