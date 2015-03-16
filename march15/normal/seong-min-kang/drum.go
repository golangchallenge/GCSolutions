// Package drum is supposed to implement the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
package drum

import "fmt"

// Pattern is the high level representation of the drum pattern
type Pattern struct {
	Version string
	Tempo   float32
	Tracks  []Tracker
}

// FormatVersion returns the string representation for the version printout
func (p Pattern) FormatVersion() string {
	return fmt.Sprintf("Saved with HW Version: %s", p.Version)
}

// FormatTempo returns the string representation for the tempo printout
func (p Pattern) FormatTempo() string {
	rep := fmt.Sprintf("%.2f", p.Tempo)
	// remove trailing zero(s)
	n := len(rep)
	for n > 0 && rep[n-1] == '0' {
		n--
	}
	if n > 0 && rep[n-1] == '.' {
		n--
	}
	s := rep[:n]
	return fmt.Sprintf("Tempo: %s", s)
}

// String returns the formatted printout of the full pattern
func (p Pattern) String() string {
	str := p.FormatVersion() + "\n"
	str += p.FormatTempo() + "\n"
	for _, tr := range p.Tracks {
		str += FormatTrackTitle(tr)
		str += "\t"
		str += FormatTrack(tr)
		str += "\n"
	}
	return str
}

// Tracker holds enough information to display a track
type Tracker interface {
	ID() uint8
	Name() string
	Beats() int
	StepsPerBeat() int
	Track() []byte
}

// FormatTrackTitle formats the track title
func FormatTrackTitle(t Tracker) string {
	return fmt.Sprintf("(%d) %s", t.ID(), t.Name())
}

// FormatTrack returns the track representation as a string
func FormatTrack(t Tracker) string {
	var str string
	for b := 0; b < t.Beats(); b++ {
		str += "|"
		for s := 0; s < t.StepsPerBeat(); s++ {
			if t.Track()[t.Beats()*b+s] == 1 {
				str += "x"
			} else {
				str += "-"
			}
		}
	}
	str += "|"
	return str
}
