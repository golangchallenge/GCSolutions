package drum

import (
	"bytes"
	"fmt"
)

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	version string  // Hardware version that the file was saved with
	tempo   float32 // Tempo of pattern in beats per minute
	tracks  []track // The actual track information
}

// String returns a human-readable string representation of the Pattern.
func (p Pattern) String() string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "Saved with HW Version: %s\n", p.version)
	fmt.Fprintf(&b, "Tempo: %g\n", p.tempo)
	for _, t := range p.tracks {
		fmt.Fprintln(&b, t)
	}
	return b.String()
}

// step indicates whether or not to trigger a sound for this step of the track.
type step uint8

// Valid step values
const (
	stepOn  = step(1) // Trigger sound
	stepOff = step(0) // Don't trigger sound
)

// String returns the string representation of a step.
func (s step) String() string {
	if s == stepOn {
		return "x"
	}
	return "-"
}

const (
	// Splice files have a fixed number of steps per track.
	stepsPerTrack = 16
	// In string representation of a track, put a '|' between groups of this many steps.
	stepsPerSeparator = 4
)

type track struct {
	id    uint8
	name  string
	steps [stepsPerTrack]step
}

// String returns a human-readable string representation of the track.
func (t track) String() string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "(%d) %s\t", t.id, t.name)
	for i, s := range t.steps {
		// Put a seperator between groups of steps.
		if i%stepsPerSeparator == 0 {
			b.WriteByte('|')
		}
		b.WriteString(s.String())
	}
	b.WriteByte('|')
	return b.String()
}
