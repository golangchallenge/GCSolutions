package drum

import (
	"bytes"
	"io"
)

var (
	notes = map[byte][]byte{
		0: []byte{'-'},
		1: []byte{'x'},
	}
	sep = []byte{'|'}
)

// Beats represents 4 measures for a single instrument.
type Beats [4]Measure

// Produces the music output, for example:
// |x---|----|x---|----|
func (b *Beats) String() string {
	buf := bytes.NewBuffer(sep)
	for _, m := range b {
		m.write(buf)
		buf.Write(sep)
	}
	return buf.String()
}

// Measure represents 4 beats in a measure.
type Measure [4]byte

// Turns []byte{0x00, 0x00, 0x00, 0x01} into "---x"
func (m *Measure) write(buf io.Writer) {
	for _, b := range m {
		buf.Write(notes[b])
	}
}
