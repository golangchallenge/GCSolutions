package drum

import (
	"fmt"
)

var (
	// ErrInvalidTrack indicates the track is in an invalid or unknown format
	ErrInvalidTrack = fmt.Errorf("Invalid Track Format")
)

// Track represents a series of steps in a measure.  A Track includes an
// integer ID, a string name and 16 steps.  A step either be silent or it can
// indicate activating the sound represented by the track (e.g. a snare drum or
// hihatt hit).
type Track struct {
	id    uint32
	name  string
	steps [16]byte
}

// Decode will read from the Reader and populate the fields
// of the track.  It returns the number of bytes read
func (t *Track) Decode(r Reader) error {
	r.Next(&t.id)
	r.VarString(&t.name)
	r.Next(&t.steps)
	// Verify the track steps
	for _, step := range t.steps {
		if step != 0x00 && step != 0x01 {
			return ErrInvalidTrack
		}
	}
	return r.Err()
}

// String returns a string representation of the entire track using the
// following format:
//
//     (<ID>) <NAME> <TAB> <MeasureString>
//
// For example
//
//     (0) kick     |x---|x---|x---|x---|
func (t *Track) String() string {
	return fmt.Sprintf("(%d) %s\t%s\n", t.id, t.name, t.MeasureString())
}

// MeasureString returns a string representation of the 16 steps in the
// measure.  For instance, a measure that activates the sound on the first, fourth,
// eighth and twelfth step would look like:
//
//     |x---|x---|x---|x---|
//
// x's represent an active sound, a dash represents silence
func (t *Track) MeasureString() string {
	s := ""
	for i, step := range t.steps {
		if i%4 == 0 {
			s += "|"
		}
		if step == 0 {
			s += "-"
		} else {
			s += "x"
		}
	}
	s += "|"
	return s
}

// Length returns the byte length of the track.  Used primarily to keep track
// of how many more bytes need to be read from the splice.  The length of a
// track is variable because the name itself is not a fixed width field in the
// splice binary format.  Therefore, the length is calculated like this:
//
// 4 bytes for the track unsigned 32 bit track ID
// 1 byte for the unsigned integer representing the length of the name
// N bytes for the name
// 16 bytes for the steps in the measure
//
func (t *Track) length() int {
	return len(t.name) + 21
}
