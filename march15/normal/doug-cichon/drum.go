// Package drum is supposed to implement the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
package drum

import (
	"bytes"
	"fmt"
)

func (p *Pattern) String() string {
	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintln("Saved with HW Version:", p.Version))
	buffer.WriteString(fmt.Sprintln("Tempo:", p.Tempo))

	for _, track := range p.Tracks {
		buffer.WriteString(fmt.Sprintln(track.String()))
	}
	return buffer.String()
}

func (t *Track) String() string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("(%d) %s\t", t.ID, t.Instrument))

	for i, beat := range t.Beats {
		if i%4 == 0 {
			buffer.WriteString("|")
		}

		if beat {
			buffer.WriteString("x")
		} else {
			buffer.WriteString("-")
		}
	}

	buffer.WriteString("|")

	return buffer.String()
}
