package drum

import (
	"bytes"
	"encoding/binary"
	"os"
	"text/template"
)

var (
	it *template.Template
)

const (
	iDoc = `({{.ID}}) {{.Name}}	{{.Beats}}`
)

func init() {
	it = template.Must(template.New("instrument").Parse(iDoc))
}

// Instrument represents a single instrument in a
// .splice file (kick, snare, etc).
type Instrument struct {
	InstrumentHeader
	Name  string
	Beats Beats
	fp    *os.File
}

func (i *Instrument) String() string {
	var buf bytes.Buffer
	it.Execute(&buf, i)
	return buf.String()
}

// InstrumentHeader represents the fixed-width data
// for an Instrument (can be loaded with binary.Read).
// PS: I don't like go lint's ID instead of Id suggestion.
type InstrumentHeader struct {
	ID     uint32
	Length uint8
}

// newInstrument takes a file pointer and creates an
// Instrument.
func newInstrument(fp *os.File) (Instrument, error) {
	i := Instrument{fp: fp}
	if err := binary.Read(i.fp, binary.LittleEndian, &i.InstrumentHeader); err != nil {
		return i, err
	}
	if err := parseName(&i); err != nil {
		return i, err
	}
	return i, binary.Read(i.fp, binary.LittleEndian, &i.Beats)
}

func parseName(i *Instrument) error {
	n := make([]byte, i.Length)
	_, err := i.fp.Read(n)
	i.Name = string(n)
	return err
}
