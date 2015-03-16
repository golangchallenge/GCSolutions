package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"strings"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
// The drum machine file consists of a header followed by the tracks.
// Header: the header has a fixed size so we use a struct to read it in one go.
// Tracks: the tracks have a variable length, therefore we loop through the
// tracks until we have read all the bytes. The length is stored in the header.
// Before reading in the tracks we create a slice that can store the maximum
// number of tracks. We also create a counter that holds the actual number of tracks.
// After looping through the tracks we re-slice the tracks slice with the actual
// number of tracks.
func DecodeFile(path string) (*Pattern, error) {
	const (
		headerSize     = 50 - 14 // total size - offset for fileSize
		fixedTrackSize = 21      // minimum number of bytes in a track (4+1+16)
	)

	type header struct {
		FileType [6]byte
		_        [7]byte // filler or FileSize is BigEndian uint64, unknown
		FileSize uint8
		Version  [32]byte
		Tempo    float32
	}

	b, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	var (
		buf = bytes.NewBuffer(b)
		h   header

		trackID         uint32
		trackNameLength uint8
		TrackSteps      [16]byte
	)

	// HEADER
	err = binary.Read(buf, binary.LittleEndian, &h)
	if err != nil {
		panic(err)
	}

	if string(h.FileType[:]) != "SPLICE" {
		return nil, fmt.Errorf("%v is not is valid splice file. (expected SPLICE, got %v)", path, string(h.FileType[:]))
	}

	var (
		maxTracks = (h.FileSize - headerSize) / fixedTrackSize
		tracks    = make([]Track, maxTracks)
		counter   int32
		steps     [16]bool
	)

	// TRACKS
	for bytesLeft := h.FileSize - headerSize; bytesLeft != 0; {

		if err := binary.Read(buf, binary.LittleEndian, &trackID); err != nil {
			return nil, fmt.Errorf("error reading track id (track #%v); %v", counter+1, err)
		}

		if err := binary.Read(buf, binary.LittleEndian, &trackNameLength); err != nil {
			return nil, fmt.Errorf("error reading track name lenght (track id %v); %v", trackID, err)
		}

		trackName := make([]byte, trackNameLength)
		if err := binary.Read(buf, binary.LittleEndian, &trackName); err != nil {
			return nil, fmt.Errorf("error reading track name (track id %v); %v", trackID, err)
		}

		if err := binary.Read(buf, binary.LittleEndian, &TrackSteps); err != nil {
			return nil, fmt.Errorf("error reading track steps lenght (track id %v); %v", trackID, err)
		}

		for i, v := range TrackSteps {
			if v == 0x00 {
				steps[i] = false
			} else if v == 0x01 {
				steps[i] = true
			} else {
				return nil, fmt.Errorf("invalid step value (track id %v, step #%v, expected 0x00 or 0x01, got %#02x)", trackID, i, v)
			}

		}

		tracks[counter] = Track{ID: trackID, Name: string(trackName[:]), Steps: steps}

		bytesLeft = bytesLeft - fixedTrackSize - uint8(trackNameLength)
		counter++
	}
	tracks = tracks[:counter] // re-slice the tracks slice to fit the actual number of tracks
	p := &Pattern{Version: strings.TrimRight(string(h.Version[:]), "\x00"), Tempo: h.Tempo, Tracks: tracks}
	return p, nil
}
