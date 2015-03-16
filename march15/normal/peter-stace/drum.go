// Package drum is supposed to implement the decoding of .splice drum machine
// files.  See golang-challenge.com/go-challenge1/ for more information
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

func (p *Pattern) String() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "Saved with HW Version: %s\n", p.Version)
	fmt.Fprintf(&buf, "Tempo: %g\n", p.Tempo)
	for _, track := range p.Tracks {
		fmt.Fprintf(&buf, "%s\n", track.String())
	}
	return buf.String()
}

// Track represents a single instrument track in a pattern.
type Track struct {
	Index    uint32
	Name     string
	Triggers [16]bool
}

func (t *Track) String() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "(%d) %s\t", t.Index, t.Name)
	for i, trigger := range t.Triggers {
		if i%4 == 0 {
			buf.WriteRune('|')
		}
		if trigger {
			buf.WriteRune('x')
		} else {
			buf.WriteRune('-')
		}
	}
	buf.WriteRune('|')
	return buf.String()
}
