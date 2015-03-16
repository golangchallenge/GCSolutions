// Package drum is supposed to implement the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
package drum

import (
	"bytes"
	"fmt"
)

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	HardwareVersion string
	Tempo           float32
	Tracks          []*Track
}

func (p *Pattern) String() string {
	// docs say str.WriteString always returns a nil error
	var str bytes.Buffer

	str.WriteString(fmt.Sprintf("Saved with HW Version: %s\nTempo: %g\n",
		p.HardwareVersion, p.Tempo))

	for _, t := range p.Tracks {
		str.WriteString(t.String())
	}

	return str.String()
}

// Track for a single instrument
type Track struct {
	ID    uint32
	Name  string
	Steps [16]bool
}

func (t *Track) String() string {
	return fmt.Sprintf("(%d) %s\t|%s|%s|%s|%s|\n",
		t.ID, t.Name, t.quarter(0), t.quarter(1), t.quarter(2), t.quarter(3))
}

// bar prints the 16th notes for the given quarter note
func (t *Track) quarter(n int) []byte {
	var str [4]byte
	for i, start := 0, n*4; i < 4; i++ {
		if t.Steps[start+i] {
			str[i] = 'x'
		} else {
			str[i] = '-'
		}
	}
	return str[:]
}
