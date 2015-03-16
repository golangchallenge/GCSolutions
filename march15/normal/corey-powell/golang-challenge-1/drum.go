// Package drum is supposed to implement the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
package drum

import (
	"fmt"
)

// Tracks are high level representation of the individual tracks present in a
// Splice/Pattern.
type Track struct {
	ID    int32
	Name  string
	Steps [16]byte
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version string
	Tempo   float32
	Tracks  []*Track
}

// Of course we need some way to export the Pattern as a string
func (p *Pattern) Export() string {
	content := "Saved with HW Version: " + p.Version + "\n"
	// some unholy hack to determine if this number needs float formatting.
	f := p.Tempo - float32(int(p.Tempo))
	if f > 0 {
		// float formatting sucks .-.
		content += fmt.Sprintf("Tempo: %.1f\n", p.Tempo)
	} else {
		content += fmt.Sprintf("Tempo: %d\n", int(p.Tempo))
	}
	for _, trk := range p.Tracks {
		content += fmt.Sprintf("(%d) %s\t", trk.ID, trk.Name)
		for j, stp := range trk.Steps {
			// This shit, is fugly, but it works
			if j%4 == 0 {
				content += "|"
			}
			if stp == 0 {
				content += "-"
			} else {
				content += "x"
			}
		}
		content += "|\n"
	}
	return content
}

func (p *Pattern) String() string {
	return p.Export()
}
