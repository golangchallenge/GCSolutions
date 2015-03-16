// Package drum is supposed to implement the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
package drum

import (
	"fmt"
	"strconv"
)

const (
	DrumVersion string = "0.616-alpha"
)

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version string
	Tempo   float32
	Tracks  []*Track
}

// Turn off all the steps of the Pattern (but keeps everything else intact).
func (p *Pattern) ClearSteps() {
	for _, t := range p.Tracks {
		for i, _ := range t.Steps {
			t.Steps[i] = false
		}
	}
}

// Get a textual representation of the Pattern.
func (p *Pattern) String() string {
	s := fmt.Sprintf("Saved with HW Version: %s\nTempo: %s\n", p.Version, strconv.FormatFloat(float64(p.Tempo), 'f', -1, 32))
	for _, t := range p.Tracks {
		s += fmt.Sprintf("%s\n", t)
	}
	return s
}

// A Track is a high level representation of a Pattern's tracks.
// Simplified, a track is an instrument, and rules for on which «steps» the instrument should make a sound.
// A step is a 16th of a bar (and thus a 4th of a beat).
type Track struct {
	Id    int32
	Name  string
	Steps []bool
}

// Get a textual representation of a Track.
func (t *Track) String() string {
	s := fmt.Sprintf("(%d) %s\t", t.Id, t.Name)

	for i, st := range t.Steps {
		if i%4 == 0 {
			s += "|"
		}

		if st {
			s += "x"
		} else {
			s += "-"
		}
	}

	return s + "|"
}
