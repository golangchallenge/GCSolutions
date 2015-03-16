// Package drum is supposed to implement the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
package drum

import "fmt"

// String return a string representation of p that is human readable
func (p Pattern) String() string {
	s := fmt.Sprintf("Saved with HW Version: %v\nTempo: %v\n", p.Version, p.Tempo)
	for i := 0; i < len(p.Tracks); i++ {
		s += fmt.Sprintf("(%v) %v\t", p.Tracks[i].ID, p.Tracks[i].Name)
		for j := 0; j < len(p.Tracks[i].Play); j++ {
			if j%4 == 0 {
				s += "|"
			}
			if p.Tracks[i].Play[j] {
				s += "x"
			} else {
				s += "-"
			}
		}
		s += "|\n"
	}
	return s
}
