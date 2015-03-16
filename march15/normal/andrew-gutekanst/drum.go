// Package drum is supposed to implement the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
package drum

import "fmt"

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version string
	Tempo   float32
	Tracks  []*Track
}

// Track represents a single drum track
type Track struct {
	ID    int
	Name  string
	Beats [16]byte
}

func (p *Pattern) String() string {
	var output string
	
	// Output Version & Tempo
	output += fmt.Sprintln("Saved with HW Version:", p.Version)
	output += fmt.Sprintln("Tempo:", p.Tempo)

	// Loop over each track, printing id, name and beats for each one
	for _, track := range p.Tracks {
		output += fmt.Sprintf("(%d) %s\t", track.ID, track.Name)

		// Print the beats by first printing a line each 4 beats,
		// then printing 'x' for a beat, '-' for a non-beat
		for k, v := range track.Beats {
			if k%4 == 0 {
				output += fmt.Sprint("|")
			}
			if v == 1 {
				output += fmt.Sprint("x")
			} else {
				output += fmt.Sprint("-")
			}
		}

		// Add a closing '|'
		output += fmt.Sprintln("|")
	}
	return output
}
