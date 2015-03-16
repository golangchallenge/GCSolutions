package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// Returns a human readable version of the Track. For example:
//	(5) cowbell	|----|----|x---|----|
func (tr *Track) String() string {
	result := fmt.Sprintf("(%d) %s\t", tr.ID, tr.Name)
	for i, val := range tr.Steps {
		// Every 4th step is a quarter note, needs a '|' delimiter.
		if i%4 == 0 {
			result += "|"
		}
		// A byte value of 1 means we play this track at this step.
		if val == 0x01 {
			result += "x"
		} else {
			result += "-"
		}
	}
	return result + "|"
}

// Track-specific decoding methods follow.

// NewTrack extracts a single track from the provided 'input'. Returns a pointer to a
// Track struct with the appropriate values, the number of bytes consumed and an error
// if applicable. See DecodeFile() in decoder.go for an overview of the track layout in
// a splice file.
func NewTrack(input io.Reader) (*Track, int, error) {
	// Initialize an empty track and offset.
	offset := 0
	var tr Track

	// Step 1: Extract the 1-byte track ID.
	id, offset, err := readBytes(input, 1, offset)
	if err != nil {
		return nil, offset, err
	}
	tr.ID = id[0]

	// Step 2: Extract the name.
	tr.Name, offset, err = readTrackName(input, offset)
	if err != nil {
		return nil, offset, err
	}

	// Step 3: Extract track steps (16 bytes).
	var steps []byte
	steps, offset, err = readBytes(input, 16, offset)
	if err != nil {
		return nil, offset, err
	}
	tr.Steps = steps

	return &tr, offset, nil
}

// Extracts the length-prefixed name of a track. Returns the
// name as a string, the new offset, and an error if applicable.
func readTrackName(input io.Reader, offset int) (string, int, error) {
	var err error
	var value []byte

	// Extract the track name length (4 bytes).
	value, offset, err = readBytes(input, 4, offset)
	if err != nil {
		return "", offset, err
	}
	var length uint32
	err = binary.Read(bytes.NewReader(value), binary.BigEndian, &length)
	if err != nil {
		return "", offset, err
	}

	// A track name length cannot be 0.
	if length == 0 {
		return "", offset, fmt.Errorf("expected non-zero track name length")
	}

	// Extract the track name (trLength bytes).
	var name []byte
	name, offset, err = readBytes(input, int(length), offset)
	if err != nil {
		return "", offset, err
	}

	return string(name), offset, nil
}

// Track-specific encoding methods follow.

// Writes the track ID, name and steps to the provided 'output'.
func (tr *Track) Write(output io.Writer) error {
	// Write the 1-byte ID.
	n, err := output.Write([]byte{tr.ID})
	if err != nil {
		return err
	}
	if n != 1 {
		return fmt.Errorf("couldn't write track ID (wrote %d bytes)", n)
	}

	// Write the track name length as 4 bytes, big-endian.
	err = binary.Write(output, binary.BigEndian, uint32(len(tr.Name)))
	if err != nil {
		return err
	}

	// Write the track name.
	n, err = output.Write([]byte(tr.Name))
	if err != nil {
		return err
	}
	if n != len(tr.Name) {
		// What about unicode track names? Ignore this check for now.
	}

	// Write the track steps.
	n, err = output.Write(tr.Steps)
	if err != nil {
		return err
	}
	if n != len(tr.Steps) {
		fmt.Errorf("couldn't write track steps (wrote %d bytes, expected %d)",
			n, len(tr.Steps))
	}

	return nil
}
