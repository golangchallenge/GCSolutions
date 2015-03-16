package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
)

// Returns a human readable version of the Pattern, including tracks.
// For example:
//	Saved with HW Version: 0.808-alpha
//	Tempo: 98.4
//	(0) kick    |x---|----|x---|----|
//	(1) snare   |----|x---|----|x---|
//	(3) hh-open |--x-|--x-|x-x-|--x-|
//	(5) cowbell |----|----|x---|----|
func (pt *Pattern) String() string {
	result := fmt.Sprintf("Saved with HW Version: %s\n", pt.Version)
	result += fmt.Sprintf("Tempo: %g\n", pt.Tempo)
	for _, track := range pt.Tracks {
		result += fmt.Sprintf("%s\n", track)
	}
	return result
}

// Pattern specific decoding methods follow.

// NewPattern decodes 'length' bytes of data from 'input' and constructs
// a Pattern with appropriate values.
func NewPattern(input io.Reader, length uint64) (*Pattern, error) {
	// Initialize empty pattern and offset.
	offset := 0
	var err error
	var pt Pattern

	// Extract version string.
	pt.Version, offset, err = readVersion(input, offset)
	if err != nil {
		return nil, err
	}

	// Extract tempo.
	pt.Tempo, offset, err = readTempo(input, offset)
	if err != nil {
		return nil, err
	}

	// Now, for all remaining bytes as per the specified length, keep reading tracks.
	for offset < int(length) {
		// Handoff to Track to handle decoding of individual tracks.
		tr, n, err := NewTrack(input)
		if err != nil {
			return nil, err
		}

		// Append the track to our pattern.
		pt.Tracks = append(pt.Tracks, tr)
		offset += n
	}

	return &pt, nil
}

// Extracts the 32-byte version string from an input reader.
// Returns the version as a trimmed string, as well as the new offset
// and an error if applicable.
func readVersion(input io.Reader, offset int) (string, int, error) {
	var err error
	var value []byte

	// Read 32 bytes.
	value, offset, err = readBytes(input, 32, offset)
	if err != nil {
		return "", offset, err
	}

	// Convert the string into a usable form. Trim all trailing NULLs.
	version := strings.Trim(string(value), "\x00")
	return version, offset, nil
}

// Extracts the 4-byte tempo for a pattern from an input reader.
// Returns the tempo as a float32, the new offset and an error
// if applicable.
func readTempo(input io.Reader, offset int) (float32, int, error) {
	var err error
	var value []byte

	// Read 4 bytes.
	value, offset, err = readBytes(input, 4, offset)
	if err != nil {
		return 0, offset, err
	}

	// Convert tempo into a usable float32. Note that floats are encoded
	// as little endian.
	var tempo float32
	err = binary.Read(bytes.NewReader(value), binary.LittleEndian, &tempo)
	if err != nil {
		return 0, offset, err
	}

	return tempo, offset, nil
}

// Pattern-specific encoding methods follow.

// Writes the version string, tempo and all tracks in the pattern
// into the provided 'output'.
func (pt *Pattern) Write(output io.Writer) error {
	// Write the version string (32 bytes).
	version := make([]byte, 32)
	copy(version, pt.Version)
	n, err := output.Write(version)
	if err != nil {
		return err
	}
	if n != len(version) {
		fmt.Errorf("couldn't write version (wrote %d bytes, expected %d)",
			n, len(version))
	}

	// Write the tempo (4 bytes, little-endian).
	err = binary.Write(output, binary.LittleEndian, pt.Tempo)
	if err != nil {
		return err
	}

	// Write each of the tracks.
	for _, track := range pt.Tracks {
		err = track.Write(output)
		if err != nil {
			return err
		}
	}

	return nil
}
