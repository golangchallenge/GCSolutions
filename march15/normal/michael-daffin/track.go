package drum

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"
)

// Step is a single step in a track
type Step uint8

// String returns a printable reprsentation of a step
func (step Step) String() string {
	if step == 0 {
		return "-"
	}
	return "x"
}

// Track contains a single track found in the .splice file
type Track struct {
	ID    uint8
	Title string
	Steps [4][4]Step
}

// Decode a track from a binrary sequence into track, returns the number of bytes read from the
// stream or an error
func (track *Track) Decode(r io.Reader) (uint, error) {
	var bytesRead uint // Number of bytes read from the stream
	if err := binary.Read(r, binary.BigEndian, &track.ID); err != nil {
		return bytesRead, err
	}
	bytesRead += 4

	var titleLength uint32
	if err := binary.Read(r, binary.BigEndian, &titleLength); err != nil {
		return bytesRead, err
	}
	bytesRead++

	title := make([]byte, titleLength)

	if err := binary.Read(r, binary.BigEndian, &title); err != nil {
		return bytesRead, err
	}
	bytesRead += uint(titleLength)

	track.Title = strings.Trim(string(title), "\x00")

	if err := binary.Read(r, binary.BigEndian, &track.Steps); err != nil {
		return bytesRead, err
	}
	bytesRead += 16

	return bytesRead, nil
}

// String returns a printable reprsentation of a track
func (track Track) String() string {
	groups := make([]string, len(track.Steps))

	for index, group := range track.Steps {
		for _, step := range group {
			groups[index] += step.String()
		}
	}

	return fmt.Sprintf("(%d) %s\t|%s|", track.ID, track.Title, strings.Join(groups, "|"))
}
