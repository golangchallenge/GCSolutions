// Package drum implements the decoding and encoding of .splice drum machine
// files. See golang-challenge.com/go-challenge1 for more information.
package drum

import (
	"bytes"
	"fmt"
)

// Pattern is the high level representation of the drum pattern contained in a
// .splice file.
type Pattern struct {
	Version string
	Tempo   float32
	Tracks  []Track

	raw []byte // trailing data from decoded pattern
}

// Track represents a single track in a Pattern.
type Track struct {
	ID    byte
	Name  string
	Steps [16]bool
}

// String returns the string representation of this Pattern.
func (p Pattern) String() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "Saved with HW Version: %s\n", p.Version)
	fmt.Fprintf(&buf, "Tempo: %v\n", p.Tempo)
	for _, track := range p.Tracks {
		fmt.Fprintln(&buf, track)
	}
	return buf.String()
}

// String returns the string representation of this Track.
func (t Track) String() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "(%d) %s\t", t.ID, t.Name)
	for i := 0; i < 16; i++ {
		if i%4 == 0 {
			fmt.Fprintf(&buf, "|")
		}
		if t.Steps[i] {
			fmt.Fprintf(&buf, "x")
		} else {
			fmt.Fprintf(&buf, "-")
		}
	}
	fmt.Fprintf(&buf, "|")
	return buf.String()
}
