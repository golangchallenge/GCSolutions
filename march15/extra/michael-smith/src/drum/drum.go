package drum

import (
	"bytes"
	"fmt"
)

// Pattern describes the drum machine pattern
type Pattern struct {
	Version string
	Tempo   float32
	Tracks  []*Track
}

// Track is one track of a pattern
type Track struct {
	ID    int32
	Name  string
	Steps Steps
}

// Steps is a set of notes in a 16 step measure
type Steps [16]bool

// FindTrack will return the first Track in a pattern with the given
// name, or nil if not found
func (p *Pattern) FindTrack(name string) *Track {
	for _, t := range p.Tracks {
		if t.Name == name {
			return t
		}
	}
	return nil
}

// AddTrack will append a new Track to the Pattern
func (p *Pattern) AddTrack(t Track) {
	p.Tracks = append(p.Tracks, &t)
}

func (p Pattern) String() string {
	var buffer bytes.Buffer

	fmt.Fprintf(&buffer, "Saved with HW Version: %s\n", p.Version)
	fmt.Fprintf(&buffer, "Tempo: %v\n", p.Tempo)

	for _, t := range p.Tracks {
		fmt.Fprintf(&buffer, "(%d) %s\t", t.ID, t.Name)
		for i, s := range t.Steps {
			if i%4 == 0 {
				buffer.WriteString("|")
			}

			if s {
				buffer.WriteString("x")
			} else {
				buffer.WriteString("-")
			}
		}
		buffer.WriteString("|\n")
	}

	return buffer.String()
}
