// Package drum is supposed to implement the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math"
)

// formatIdentifier is the binary equivalent to "SPLICE".
var formatIdentifier = []byte{0x53, 0x50, 0x4c, 0x49, 0x43, 0x45}

// UnmarshalBinary decodes the data given to the Pattern.
func (p *Pattern) UnmarshalBinary(data []byte) error {
	if len(data) == 0 {
		return errors.New("Pattern.UnmarshalBinary: no data")
	}

	expectedLength := binary.BigEndian.Uint64(data[6:14])
	dataLength := uint64(len(data))
	switch {
	case expectedLength > dataLength:
		return errors.New("Pattern.UnmarshalBinary: file too short, your data may be corrupt")
	case expectedLength < dataLength:
		data = data[:expectedLength]
	}

	f := readDataTillNull(data[:6])
	if !bytes.Equal(f, formatIdentifier) {
		return errors.New("Pattern.UnmarshalBinary: unsupported file format")
	}

	v := readDataTillNull(data[14:46])
	p.Version = string(v)

	t := binary.LittleEndian.Uint32(data[46:50])
	p.Tempo = math.Float32frombits(t)

	if err := p.readTracks(data[50:]); err != nil {
		return err
	}

	return nil
}

func (p *Pattern) readTracks(data []byte) error {
	for {
		id := binary.LittleEndian.Uint32(data[:4])
		nameEndIndex := 4 + 1 + int(data[4])

		p.Tracks = append(p.Tracks, Track{
			ID:    int32(id),
			Name:  string(data[5:nameEndIndex]),
			Steps: data[nameEndIndex : nameEndIndex+16],
		})

		if len(data) > int(nameEndIndex+16) {
			data = data[nameEndIndex+16:]
		} else {
			// There are no more tracks left in the data set.
			break
		}
	}

	return nil
}

// Read data till the first null value is found and return the data read to that point
func readDataTillNull(data []byte) []byte {
	var result = make([]byte, 0)
	for _, val := range data {
		if val == 0 {
			break
		}
		result = append(result, val)
	}

	return result
}
