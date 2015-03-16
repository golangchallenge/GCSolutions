// Package drum is supposed to implement the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
package drum

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"
)

const stepsPerMeasure = 16

// Track is the high level representation of the
// track pattern contained in a .splice file.
type Track struct {
	Id      uint32
	NameLen uint8
	Name    string
	Steps   [stepsPerMeasure]byte
}

// UnmarshalBinary unmarshalls a track from bytes within an io.Reader.
func (t *Track) UnmarshalBinary(src io.Reader) error {
	if err := binary.Read(src, binary.LittleEndian, &t.Id); err != nil {
		return fmt.Errorf("error unmarshalling id of track: %v", err)
	}

	if err := binary.Read(src, binary.LittleEndian, &t.NameLen); err != nil {
		return fmt.Errorf("error unmarshalling length of track's name: %v", err)
	}

	nameBuf := make([]byte, t.NameLen)
	if _, err := src.Read(nameBuf); err != nil {
		return fmt.Errorf("error unmarshalling name of track: %v", err)
	}
	t.Name = string(nameBuf)

	if err := binary.Read(src, binary.LittleEndian, &t.Steps); err != nil {
		return fmt.Errorf("error unmarshalling steps of track: %v", err)
	}

	return nil
}

// String returns a formatter track containing its id, name, and a full measure.
func (t *Track) String() string {
	const sep = "|"
	const beat = "x"
	const rest = "-"
	const stepsPerQuarter = 4

	notes := make([]string, len(t.Steps)/stepsPerQuarter)
	for i, step := range t.Steps {
		if step > 0 {
			notes[i/stepsPerQuarter] += beat
		} else {
			notes[i/stepsPerQuarter] += rest
		}
	}
	measure := sep + strings.Join(notes, sep) + sep

	return fmt.Sprintf("(%d) %s\t%s", t.Id, t.Name, measure)
}
