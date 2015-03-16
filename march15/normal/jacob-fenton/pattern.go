// Package drum provides an API for decoding drum patterns
// contained within .splice files.
package drum

import "fmt"

const tempoOffset = 32

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version string
	Tempo   float32

	// Each *Track describes a single track in the pattern.
	Tracks []*Track
}

func (p *Pattern) String() string {
	str := fmt.Sprintf("Saved with HW Version: %s\nTempo: %g\n",
		p.Version, p.Tempo)

	for _, track := range p.Tracks {
		str += fmt.Sprint(track)
	}

	return str
}
