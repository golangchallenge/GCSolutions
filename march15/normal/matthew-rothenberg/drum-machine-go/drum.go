// Package drum implements the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
//
// This implementation submitted by Matthew Rothenberg <http://github.com/mroth>
package drum

import (
	"bytes"
	"fmt"
)

// Pattern is the high level representation of the drum pattern contained in
// a .splice file.
type Pattern struct {
	Version string
	Tempo   float32
	Tracks  []*Track
}

func (p *Pattern) String() string {
	var output bytes.Buffer
	output.WriteString(p.versionString())
	output.WriteString(p.tempoString())

	for _, t := range p.Tracks {
		output.WriteString(t.String())
	}
	return output.String()
}

func (p *Pattern) versionString() string {
	return fmt.Sprintf("Saved with HW Version: %s\n", p.Version)
}

func (p *Pattern) tempoString() string {
	return fmt.Sprintf("Tempo: %v\n", p.Tempo)
}

// Track is the high level representation of a single instrument within a
// .splice file Pattern.
type Track struct {
	ID    uint8
	Name  string   // the name of the instrument, or "sample"
	Beats [16]Beat // grid of beat step information in 4/4 time
}

func (t *Track) String() string {
	return fmt.Sprintf("(%d) %s\t%s\n", t.ID, t.Name, t.beatGrid())
}

func (t *Track) beatGrid() string {
	var output bytes.Buffer
	output.WriteString("|")
	for i, b := range t.Beats {
		output.WriteString(b.String())
		if (i+1)%4 == 0 {
			output.WriteString("|")
		}
	}
	return output.String()
}

// Beat represents whether an instrument is "played" for a particular step
type Beat bool

func (b Beat) String() string {
	if b {
		return "x"
	}
	return "-"
}
