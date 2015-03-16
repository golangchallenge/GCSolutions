// Package drum implements the decoding of .splice drum machine files.
// See http://golang-challenge.com/go-challenge1/ for more information
package drum

import (
	"bytes"
	"fmt"
)

// Pattern is the high level representation of the drum pattern contained in a .splice file.
type Pattern struct {
	Version string
	Tempo   float32 // speed of the playback, in bpm
	Tracks  []Track
}

func (p Pattern) String() string {
	var b bytes.Buffer
	for _, t := range p.Tracks {
		b.WriteString(t.String())
	}
	return fmt.Sprintf("Saved with HW Version: %s\nTempo: %v\n%s", p.Version, p.Tempo, b.String())
}

// Track represents the sound of an instrument.
type Track struct {
	ID    uint8
	Name  string
	Steps [16]bool // if a step is true, it triggers a sound
}

func (t Track) String() string {
	var b bytes.Buffer
	for i, v := range t.Steps {
		if i%4 == 0 {
			b.WriteString("|")
		}
		if v {
			b.WriteString("x")
		} else {
			b.WriteString("-")
		}
	}
	b.WriteString("|")
	return fmt.Sprintf("(%d) %s\t%s\n", t.ID, t.Name, b.String())
}
