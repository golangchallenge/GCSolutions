// Package drum implements the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information.

// For more information regarding the .splice files,
// please see the SPLICE.md file.

package drum

import (
	"fmt"
)

var (
	// Valid version strings
	sevenZeroEightAlpha = "0.708-alpha"
	eightZeroEightAlpha = "0.808-alpha"
	nineZeroEightAlpha  = "0.909"
)

// Track is a high level representation of a single track in a Pattern.
type Track struct {
	name  string
	steps [16]bool
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	version    string
	tempo      float32
	tracks     map[int]Track // maps id to corresponding Track
	printOrder []int
}

// NewPattern returns a new Pattern with a default print order of the Tracks.
func NewPattern(version string, tempo float32, tracks map[int]Track) (p *Pattern) {
	p = new(Pattern)
	p.version = version
	p.tempo = tempo
	p.tracks = tracks
	p.makeDefaultPrintOrder()
	return p
}

// String returns a string representation of the pattern along with its tracks
// according to the Pattern's printOrder.
func (p *Pattern) String() string {
	out := fmt.Sprintf("Saved with HW Version: %s\nTempo: %g\n",
		p.version,
		p.tempo)

	if p.printOrder == nil {
		p.makeDefaultPrintOrder()
	}

	for _, id := range p.printOrder {
		track := p.tracks[id]
		out += fmt.Sprintf("(%d) %4s\t|", id, track.name)

		for i := 0; i < len(track.steps); i++ {
			if track.steps[i] {
				out += "x"
			} else {
				out += "-"
			}
			if i%4 == 3 { // Adds the vertical bar after every fourth step
				out += "|"
			}
		}
		out += "\n"
	}
	return out
}

// adds a default printOrder of the Tracks.
func (p *Pattern) makeDefaultPrintOrder() {
	p.printOrder = make([]int, 0)
	for id := range p.tracks {
		p.printOrder = append(p.printOrder, id)
	}
}
