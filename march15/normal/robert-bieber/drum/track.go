package drum

import (
	"encoding/binary"
	"io"
)

// readTrack reads a single track from the file, returning a pointer
// to a track struct or nil and an error.
func readTrack(fin io.Reader) (*Track, error) {
	t := &Track{}

	// 4 bytes for ID, little-endian
	err := binary.Read(fin, binary.LittleEndian, &t.ID)
	if err != nil {
		return nil, err
	}

	// One byte for name length
	var nameLen uint8
	err = binary.Read(fin, binary.LittleEndian, &nameLen)
	if err != nil {
		return nil, err
	}
	nameBytes := make([]byte, nameLen)
	n, err := fin.Read(nameBytes)
	if n != int(nameLen) || err != nil {
		return nil, err
	}
	t.Name = string(nameBytes)

	// 16 bytes for beat data, array of individual bytes
	beatData := make([]byte, sixteenthsPerMeasure)
	n, err = fin.Read(beatData)
	if n != sixteenthsPerMeasure || err != nil {
		return nil, err
	}
	for k, v := range beatData {
		if v == 0 {
			t.Data[k] = Inactive
		} else {
			t.Data[k] = Active
		}
	}

	return t, nil
}
