// Package drum is supposed to implement the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
package drum

import (
	"fmt"
	"strings"
)

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version string
	Tempo   float32
	Tracks  []*Track
}

func (p *Pattern) String() string {
	ts := make([]string, 0, len(p.Tracks))
	for _, t := range p.Tracks {
		ts = append(ts, t.String())
	}
	return fmt.Sprintf(`Saved with HW Version: %s
Tempo: %g
%s
`, p.Version, p.Tempo, strings.Join(ts, "\n"))
}

// Track is the high level representation of the
// drum track.
type Track struct {
	Name string
	ID   uint32
	// Steps represents 16 steps (one bit for each step).
	Steps uint16
}

func (t *Track) String() string {
	var s [4]string // 4 quarter notes
	for i := 0; i < 4; i++ {
		var b [4]byte
		for j := 0; j < 4; j++ {
			m := uint16(1) << uint16(i*4+j)
			if t.Steps&m == m {
				b[j] = 'x'
			} else {
				b[j] = '-'
			}
		}
		s[i] = string(b[:])
	}
	return fmt.Sprintf("(%d) %s\t|%s|", t.ID, t.Name, strings.Join(s[:], "|"))
}
