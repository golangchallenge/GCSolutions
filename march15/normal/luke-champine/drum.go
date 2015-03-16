// Package drum is supposed to implement the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information.
//
// .splice files are encoded in binary. They consist of a header (containing
// the version and tempo, among other things) followed by a series of tracks.
// The exact binary format is as follows:
//
// Header:
//
//   | field description                                | field type              |
//   |--------------------------------------------------|-------------------------|
//   | The literal string "SPLICE"                      | [6]byte                 |
//   | Pattern length (in bytes)                        | int64 (big endian)      |
//   | Version string (padded with zeros)               | [32]byte                |
//   | Tempo                                            | float32 (little endian) |
//
// Track:
//
//   | field description                                | field type              |
//   |--------------------------------------------------|-------------------------|
//   | ID                                               | uint8                   |
//   | Length of name                                   | int32 (big endian)      |
//   | Name                                             | string                  |
//   | Steps, where the byte 0x01 indicates a note      | [16]byte                |
package drum

import (
	"fmt"
)

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	version string
	tempo   float32
	tracks  []track
}

const stepsPerTrack = 16

// track is the representation of a single track in a Pattern.
type track struct {
	id    uint8
	name  string
	steps [stepsPerTrack]bool
}

// String implements the fmt.Stringer interface, allowing tracks to be
// printed.
func (t track) String() string {
	// write the header
	header := fmt.Sprintf("(%d) %s\t", t.id, t.name)
	// write the steps
	steps := []byte("|----|----|----|----|")
	// add an 'x' for each note
	for i, x := range t.steps {
		if x {
			// need to adjust 'i' to account for the '|'s
			steps[i+i/4+1] = 'x'
		}
	}
	return header + string(steps)
}

// String implements the fmt.Stringer interface, allowing Patterns to be
// printed.
func (p Pattern) String() string {
	// write the header
	str := fmt.Sprintf("Saved with HW Version: %s\nTempo: %g\n", p.version, p.tempo)
	// write each track
	for _, track := range p.tracks {
		str += fmt.Sprintln(track)
	}

	return str
}
