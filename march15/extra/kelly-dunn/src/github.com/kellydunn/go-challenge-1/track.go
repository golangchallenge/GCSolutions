package drum

import (
	"encoding/binary"
	"io"
)

// Track describes a single track that is encoded inside of the splice file.
// It is comprised of an ID, a Name, and a Step Sequence which describes
// When the track should be triggered for output.
//
// Note that the Buffer and Playhead attributes are used for audio playback.
// These are intended to be loaded by and manipulated by an external library.
type Track struct {
	ID           uint8
	Name         string
	StepSequence StepSequence

	// Note:  These attributes are intended to be used
	//        for the "extra functionality" of providing audio playback.
	//        see the github.com/kellydunn/go-step-sequencer project for more details.
	Buffer   []float32
	Playhead int
}

// Reads the track Id from the passed in reader and applies to the passed in track pointer.
// Returns the number of bytes read, or an error if there is one that is encountered.
func readTrackID(reader io.Reader, t *Track) (int, error) {
	var id uint8
	err := binary.Read(reader, binary.BigEndian, &id)
	if err != nil {
		return 0, err
	}

	t.ID = id

	return TrackIDSize, nil
}

// Reads the track Step Sequence from the passed in reader and applies to the passed in track pointer.
// Returns the number of bytes read, or an error if there is one that is encountered.
func readTrackStepSequence(reader io.Reader, t *Track) (int, error) {
	steps := make([]byte, StepSequenceSize)
	err := binary.Read(reader, binary.BigEndian, &steps)
	if err != nil {
		return 0, err
	}

	t.StepSequence = StepSequence{Steps: steps}

	return StepSequenceSize, nil
}

// Reads the track Name from the passed in reader and applies to the passed in track pointer.
// Returns the number of bytes read, or an error if there is one that is encountered.
func readTrackName(reader io.Reader, t *Track) (int, error) {
	bytesRead := 0

	var trackNameLen uint32
	err := binary.Read(reader, binary.BigEndian, &trackNameLen)
	if err != nil {
		return bytesRead, err
	}

	bytesRead += TrackNameSize

	trackNameBytes := make([]byte, trackNameLen)
	err = binary.Read(reader, binary.BigEndian, trackNameBytes)
	if err != nil {
		return bytesRead, err
	}

	t.Name = string(trackNameBytes)

	bytesRead += int(trackNameLen)

	return bytesRead, nil
}
