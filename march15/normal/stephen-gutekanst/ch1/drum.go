// Package drum is supposed to implement the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
package drum

import (
	"bytes"
	"fmt"
)

// Track represents a single drum track.
type Track struct {
	ID    int
	Name  string
	Steps [16]bool
}

func (t Track) String() string {
	b := bytes.NewBuffer(nil)
	fmt.Fprintf(b, "(%d) %s\t", t.ID, t.Name)
	for q := 0; q < 4; q++ {
		fmt.Fprintf(b, "|")
		for s := q * 4; s < (q*4)+4; s++ {
			if t.Steps[s] {
				fmt.Fprintf(b, "x")
			} else {
				fmt.Fprintf(b, "-")
			}
		}
	}
	fmt.Fprintf(b, "|")
	return b.String()
}

// Pattern is the high level representation of the drum pattern contained in a
// .splice file.
//
// It has a tempo (BPM) and a series of tracks that each utilize their own
// instrument.
type Pattern struct {
	HWVersion string
	Tempo     float32
	Tracks    []Track
}

func (p *Pattern) String() string {
	// Print basic info about the pattern now.
	b := bytes.NewBuffer(nil)
	fmt.Fprintf(b, "Saved with HW Version: %s\n", p.HWVersion)
	fmt.Fprintf(b, "Tempo: %g\n", p.Tempo)

	// Print each track's info now.
	for _, track := range p.Tracks {
		fmt.Fprintf(b, "%s\n", track.String())
	}
	return b.String()
}
