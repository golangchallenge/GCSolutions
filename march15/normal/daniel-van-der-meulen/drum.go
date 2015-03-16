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
	Version string
	Tempo   float32
	Tracks  []Track
}

// String returns the string representation of a Pattern, e.g.:
// Saved with HW Version: 0.808-alpha
// Tempo: 98.4
// (0) kick    |x---|----|x---|----|
// (1) snare   |----|x---|----|x---|
// (3) hh-open |--x-|--x-|x-x-|--x-|
// (5) cowbell |----|----|x---|----|
func (p Pattern) String() string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintln("Saved with HW Version:", p.Version))
	buf.WriteString(fmt.Sprintln("Tempo:", p.Tempo))
	for _, v := range p.Tracks {
		buf.WriteString(fmt.Sprintln(v))
	}
	return buf.String()
}

// Track is the high level representation of a single sound track.
type Track struct {
	ID    uint32
	Name  string
	Steps [16]bool
}

// String returns the output of one track, e.g.: "(40) kick |x--x|----|x---|x--x|"
func (t Track) String() string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("(%v) %v\t", t.ID, t.Name))
	for i, v := range t.Steps {
		if i%4 == 0 {
			buf.WriteString("|")
		}
		if v {
			buf.WriteString("x")
		} else {
			buf.WriteString("-")
		}
	}
	buf.WriteString("|")
	return buf.String()
}
