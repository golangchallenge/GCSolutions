package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// A Track is a representation, representin an instrument
// and an indication of whether the instrument should be
// played at a given step.
type Track struct {
	Instrument Instrument // The instrument used for this track
	Steps      []Step     // Individual steps
}

// A Step indicated what action should be taken at a given step
type Step byte

const (
	// StepNothing indicates that nothing should happend this step
	StepNothing Step = iota
	// StepPlay indicated that the instrument should be played this step
	StepPlay
)

// Fixed track length in SPLICE format.
const SpliceTrackLength = 16

// Returns a human readable representation of the track.
func (t Track) String() string {
	steps := ""
	for i, step := range t.Steps {
		// Add 'Bar' every 4 beats indicated with a pipe
		if i%4 == 0 {
			steps += "|"
		}
		if step == StepNothing {
			steps += "-"
		} else {
			steps += "x"
		}
	}
	// End with a bar, if length is multiple of 4.
	if len(t.Steps)%4 == 0 {
		steps += "|"
	}
	return fmt.Sprintf("%v\t%s", t.Instrument, steps)
}

// decodeTrack will decode a single track from the supplied buffer,
// and return it.
// The buffer is forwarded to the end of the track data.
// If not more tracks can be found io.EOF will be returned.
func decodeTrack(r *bytes.Buffer) (*Track, error) {
	t := &Track{}

	// Read track header
	th := trackHeader{}
	err := binary.Read(r, binary.BigEndian, &th)
	if err != nil {
		return nil, err
	}

	t.Instrument.ID = th.ID
	err = t.Instrument.decodeName(r, int(th.NameLen))
	if err != nil {
		return nil, err
	}

	// Read beats
	t.Steps = make([]Step, SpliceTrackLength)
	for i := range t.Steps {
		c, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		t.Steps[i] = Step(c)
	}
	return t, nil
}

// Track header used for decoding.
type trackHeader struct {
	// ID of the instrument
	ID uint8
	// Lenght of the string indicating the name of the instrument.
	NameLen uint32
}

// Len returns the number beats in this track.
func (t Track) Len() int {
	return len(t.Steps)
}

// At returns the step value at x
// The value loops after the length of the tracks.
// Negative values of x will always return StepNothing
func (t Track) At(x int) Step {
	if x < 0 {
		return StepNothing
	}
	return t.Steps[x%len(t.Steps)]
}
