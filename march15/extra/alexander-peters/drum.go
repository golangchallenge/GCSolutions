// Package drum is supposed to implement the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
package drum

const (
	maxVersionLength = 32
	stepsLength      = 16
)

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	version string
	tempo   float32
	tracks  []*Track
}

// A Track represents an audio sample loaded by the drum machine,
// allowing the programmer to schedule the playback of the sound.
// The scheduling of the playback is done using the concept of steps.
type Track struct {
	id    uint32
	name  string
	steps Steps
}

// Steps are one of the parts of the measure that are being programmed
// (the programmed measure is known as a pattern). The measure (also called a bar)
// is divided in Steps.
// The drum machine only supports 16 step measure patterns played in 4/4 time.
// The measure is comprised of 4 quarter notes, each quarter note is comprised
// of 4 sixteenth notes and each sixteenth note corresponds to a step.
type Steps [stepsLength]bool
