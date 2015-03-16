package drum

import (
	"bytes"
	"strconv"
)

// Pattern represents the drum pattern contained in a .splice file.
type Pattern struct {
	Version []byte  // The hardware version.
	Tempo   float64 // The drum kit tempo as BPM.
	Tracks  []Track // The tracks that make up the pattern.
}

// String returns a representation of the Pattern in the following form:
//  Saved with HW Version: {Version}
//  Tempo: {Tempo}
//  ({ID}) {Name}\t|{QuarterNote}|{QuarterNote}|{QuarterNote}|{QuarterNote}|
//  ({ID}) {Name}\t|{QuarterNote}|{QuarterNote}|{QuarterNote}|{QuarterNote}|
//  ...
func (p Pattern) String() string {
	buf := bytes.NewBufferString("Saved with HW Version: ")
	buf.Write(p.Version)
	buf.WriteByte('\n')

	buf.WriteString("Tempo: ")
	buf.WriteString(strconv.FormatFloat(p.Tempo, 'f', -1, 32))
	buf.WriteByte('\n')

	for _, t := range p.Tracks {
		buf.WriteString(t.String())
		buf.WriteByte('\n')
	}

	return buf.String()
}
