/*
Package drum implements the decoding of .splice drum machine files.
See golang-challenge.com/go-challenge1/ for more information
This file contains the public API of the decoder.
*/
package drum

import (
	"errors"
	"os"
)

// Possible errors arrising during file decoding:

// An InvalidFile error indicates the file is not a SPLICE file.
var InvalidFile error = errors.New("file does not appear to be a SPLICE file")

// A FileTooShort error indicates the file did not contain as many bytes as was specified in the header
var FileTooShort error = errors.New("file not long enough")

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	p := &Pattern{}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	head, remaining, err := readHeader(file)
	if err != nil {
		return nil, err
	}
	if remaining < 0 {
		return nil, FileTooShort
	}
	p.Header = *head

	for {
		track, err := readTrack(file, &remaining)
		if err != nil {
			return nil, err
		}
		p.Tracks = append(p.Tracks, *track)
		if remaining == 0 {
			break
		}
	}

	return p, nil
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Header Header
	Tracks []Track
}

// Header contains the interesting information from the file header.
type Header struct {
	HardwareVersion string  // The hardware version as a UTF-8 string
	Tempo           float32 // The tempo
}

// Track contains the high level representation of a track from a file.
type Track struct {
	Id   uint32    // The instrument id number
	Name string    // The name of the instrument as a UTF-8 string
	Data TrackData // The data read from the track.
}

// TrackDataLength is the length of the track data to read in bytes.
// We expect 16 bytes of track data, as the drum machine uses 16th notes.
const TrackDataLength = 16

// TrackData is the data read from a track.
// This is not in any way changed from how it was read from the file.
// The assumption is that it will be 0 for no sound, 1 for sound.
type TrackData [TrackDataLength]byte
