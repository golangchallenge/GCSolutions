// Package drum is supposed to implement the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
package drum

import (
	"bytes"
	"fmt"
)

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version string
	Tempo   float32
	Tracks  []Track
}

// Track is the high level representation of each of the
// tracks within a Pattern
type Track struct {
	ID    int
	Name  string
	Steps [16]Step
}

// Step is the high level representation of each beat
// of a Track.
type Step struct {
	Active bool
}

// Default string representation of Pattern
func (p Pattern) String() string {
	var buffer bytes.Buffer

	var template = "Saved with HW Version: %s\nTempo: %v"
	buffer.WriteString(fmt.Sprintf(template, p.Version, p.Tempo))
	for _, track := range p.Tracks {
		buffer.WriteString(fmt.Sprintf("\n%v", track))
	}
	buffer.WriteString("\n")
	return buffer.String()
}

func (t Track) String() string {
	var buffer bytes.Buffer
	var template = "(%d) %s\t"
	buffer.WriteString(fmt.Sprintf(template, t.ID, t.Name))
	for i, step := range t.Steps {
		if i%4 == 0 {
			buffer.WriteString("|")
		}
		buffer.WriteString(fmt.Sprintf("%v", step))
	}
	buffer.WriteString("|")
	return buffer.String()
}

func (s Step) String() string {
	if s.Active {
		return "x"
	}

	return "-"
}
