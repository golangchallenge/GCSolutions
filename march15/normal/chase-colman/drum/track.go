package drum

import (
	"encoding/binary"
	"fmt"
	"io"
)

// QuarterNote represents 4 sixteenth notes as a byte with a value of 1 or 0
type QuarterNote [4]byte

// String returns a representation of the QuarterNote as it should
// appear in a drum kit measure. The sixteenth note steps are marked with 'x'.
// Empty steps are marked with '-'.
func (n QuarterNote) String() string {
	res := []byte{'-', '-', '-', '-'}
	for i := range n {
		if n[i] == 1 {
			res[i] = 'x'
		}
	}
	return string(res)
}

// Track represents a drum kit track with an ID, instrument name,
// and a measure of 4 quarter notes.
type Track struct {
	ID      uint32
	Name    []byte         // Instrument name
	Measure [4]QuarterNote // 4 quarter notes, 16 sixteenth notes
}

// ReadTrack reads 4 bytes as a little endian uint32 to retrieve a track ID,
// reads a byte for a string length, reads that length to retrieve
// the track's instrument name, then reads 16 bytes as the tracks measure.
// It returns the resulting reads as a drum kit Track and an error, if any.
func ReadTrack(r io.Reader) (*Track, error) {
	t := Track{}

	err := binary.Read(r, binary.LittleEndian, &t.ID)
	if err != nil {
		return nil, err
	}

	nameLen := []byte{0}
	_, err = io.ReadFull(r, nameLen)
	if err != nil {
		return nil, err
	}

	t.Name = make([]byte, int(nameLen[0]))
	_, err = io.ReadFull(r, t.Name)
	if err != nil {
		return nil, err
	}

	measure := make([]byte, 16)
	_, err = io.ReadFull(r, measure)
	if err != nil {
		return nil, err
	}

	for i := range t.Measure {
		t.Measure[i][0] = measure[4*i+0]
		t.Measure[i][1] = measure[4*i+1]
		t.Measure[i][2] = measure[4*i+2]
		t.Measure[i][3] = measure[4*i+3]
	}

	return &t, nil
}

// String returns a representation of the Track as in the following form:
//  ({ID}) {Name}\t|{QuarterNote}|{QuarterNote}|{QuarterNote}|{QuarterNote}|
// The sixteenth note steps of each QuarterNote are marked with 'x'.
// Empty steps are marked with '-'.
func (t Track) String() string {
	return fmt.Sprintf(
		"(%d) %s\t|%s|%s|%s|%s|",
		t.ID, t.Name,
		t.Measure[0], t.Measure[1], t.Measure[2], t.Measure[3],
	)
}
