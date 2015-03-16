// Package drum implements a simple library for Splice files.
// A Splice file typically ends with an extencion ".splice"
package drum

import (
	"fmt"
)

// Prints Pattern, *p, in a human-readable format
func (p *Pattern) String() string {
	output := fmt.Sprintln("Saved with HW Version:", p.Hardware)
	output = output + fmt.Sprintln("Tempo:", p.Tempo)
	for _, track := range p.Tracks {
		output = output + fmt.Sprintln(track)
	}
	return output
}

// Prints Track, *t, in a human-readable format
func (t Track) String() string {
	output := fmt.Sprintf("(%d) %s\t", t.ID, t.Name)
	display := t.toDisplayForm()
	output = output + fmt.Sprintf("|%s|%s|%s|%s|", display[0:4], display[4:8], display[8:12], display[12:16])
	return output
}

// from Track Steps (t.Steps), create a [16]byte ready for printing
// t.Steps are stored as binary 0 & 1's
//	 ie. [16]byte 1001100100001111
// returns a byte array filled with '-' & 'x's
//   ie. [16]byte x--xx--x----xxxx
func (t Track) toDisplayForm() [16]byte {
	var display [16]byte
	for i, v := range t.Steps {
		if v == 0 {
			display[i] = '-'
		} else if v == 1 {
			display[i] = 'x'
		} else {
			display[i] = 'e'
		}
	}
	return display
}

// PlayStep returns true if step, s, of Track t needs to be played (equals 1)
// Otherwise, returns false
func (t Track) PlayNote(s int) bool {
	if t.Steps[s] == 1 {
		return true
	}
	return false
}

