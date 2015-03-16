package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// DecodeFile decodes the .splice file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
//
// A drum file is binary and has the following format:
//	6 bytes  : The word "SPLICE" in ASCII.
//	8 bytes  : A big endian integer describing the length (in bytes) of the remaining data.
//	32 bytes : A NULL-terminated version string.
//	4 bytes  : A little endian single-precision floating point describing the tempo.
//
// All following bytes consist of a series of tracks. A track is laid out as:
//	1 byte   : The ID of the track.
//	n bytes  : A length prefixed string that is the name of the track. The length is prefixed as a 4 byte big endian integer.
//	16 bytes : Each of the 16 steps for this track, 1 byte each.
func DecodeFile(path string) (*Pattern, error) {
	// Open our splice file (and defer close).
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Read and validate the header.
	offset, err := readHeader(f)
	if err != nil {
		return nil, err
	}

	// Extract length of the remaining data.
	var remaining uint64
	remaining, offset, err = readLength(f, offset)
	if err != nil {
		return nil, err
	}

	// For now, we assume that there is a single pattern per splice file.
	// (This is not true for pattern_5.splice!). Handoff to the Pattern
	// to read the remaining data.
	pt, err := NewPattern(f, remaining)
	if err != nil {
		return nil, err
	}

	return pt, nil
}

// This is a generic helper method. It reads 'n' bytes from reader 'input'.
// Returns the bytes read as a slice, offset + n, and an error.
func readBytes(input io.Reader, n int, offset int) ([]byte, int, error) {
	value := make([]byte, n)
	rd, err := input.Read(value)
	if err != nil {
		return nil, offset, err
	}
	if rd != n {
		return nil, offset, fmt.Errorf("expected to read %d bytes, got %d", n, rd)
	}
	return value, offset + rd, nil
}

// Extracts a 6-byte long header and validates it. Returns the new offset
// and an error if applicable.
func readHeader(f *os.File) (int, error) {
	// Read 6 bytes.
	expected := "SPLICE"
	value, _, err := readBytes(f, len(expected), 0)
	if err != nil {
		return 0, err
	}

	// Validate header value.
	if string(value) != expected {
		return 0, fmt.Errorf("corrupt or missing header (%s)", value)
	}

	return len(expected), nil
}

// Extracts the 8-byte long splice file length, encoded as a big-endian integer.
// Returns the length as a uint64, the new offset and an error if applicable.
func readLength(f *os.File, offset int) (uint64, int, error) {
	var err error
	var value []byte

	// Read 8 bytes.
	value, offset, err = readBytes(f, 8, offset)
	if err != nil {
		return 0, offset, err
	}

	// Convert value into a uint64.
	var length uint64
	err = binary.Read(bytes.NewReader(value), binary.BigEndian, &length)
	if err != nil {
		return 0, offset, err
	}

	// We could, at this stage, validate the length. However, not all splice
	// file are guaranteed to be of this length. In particular, pattern_5.splice
	// seems to have another file appended at the end, which we will just ignore.
	// Let's just check that the length is non-zero instead.
	if length == 0 {
		return 0, offset, fmt.Errorf("expected non-zero content length")
	}

	return length, offset, nil
}
