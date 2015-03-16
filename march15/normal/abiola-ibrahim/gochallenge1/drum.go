// Package drum is supposed to implement the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
package drum

import "fmt"

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version string
	Tempo   float32
	Tracks  Tracks
}

// String gets the human readable format of the Pattern in a .splice file.
func (p *Pattern) String() string {
	str := fmt.Sprintf("Saved with HW Version: %v\nTempo: %v\n", p.Version, p.Tempo)
	str += p.Tracks.String()
	return str
}

// Track is a high level representation of track. There are many tracks in Pattern
type Track struct {
	Id    byte
	Name  string
	Steps [16]bool
}

// String gets the human readable format of the track
func (track *Track) String() string {
	str := fmt.Sprintf("(%v) %v", track.Id, track.Name)
	d := 16 - len(track.Name) - len(fmt.Sprintf("(%v) ", track.Id))
	tabSpace := d / 8
	if d%8 > 0 {
		tabSpace++
	}
	for i := 0; i < tabSpace; i++ {
		str += "\t"
	}
	str += "|"
	for i, s := range track.Steps {
		if s {
			str += "x"
		} else {
			str += "-"
		}
		if (i+1)%4 == 0 {
			str += "|"
		}
	}
	return str
}

// Tracks is declared to make []Track satisfy the Stringer interface
type Tracks []Track

// String gets the human readable format of all tracks in the underlying slice
func (tracks *Tracks) String() string {
	str := ""
	for i := 0; i < len(*tracks); i++ {
		str += fmt.Sprintln((*tracks)[i].String())
	}
	return str
}
