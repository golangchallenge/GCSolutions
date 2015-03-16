package drum

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const staticBytesPerTrack = 0x15

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	_, fileName := filepath.Split(path)
	p := &Pattern{
		FileName: fileName,
		Tracks:   make([]*Track, 0),
	}

	fin, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("drum: Error opening file %s", path)
	}
	defer fin.Close()

	remainingBytes, err := readHeader(fin, p)
	if err != nil {
		return nil, errors.New("drum: Error reading file header")
	}

	for remainingBytes > 0 {
		newTrack, err := readTrack(fin)
		if err != nil {
			return nil, fmt.Errorf(
				"drum: Error reading track %d",
				len(p.Tracks),
			)
		}
		p.Tracks = append(p.Tracks, newTrack)
		remainingBytes -= staticBytesPerTrack
		remainingBytes -= int32(len(newTrack.Name))
	}

	return p, nil
}
