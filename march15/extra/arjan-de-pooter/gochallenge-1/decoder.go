package drum

import (
	"fmt"
	"os"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
// Byte offsets:
// 0, 6: File header string: SPLICE
// 6, 8: Content size int64
// 14, 32: Version string
// 46, 4: Tempo float
// 50, size - 36: []Tracks
// Track byte offsets:
// 0, 4: ID int32
// 4, 1: length of track name int8
// 5, length: track name string
// 5 + length, 16: steps 00 or 01
func DecodeFile(path string) (*Pattern, error) {
	p := &Pattern{}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	header, err := readHeader(f)
	if err != nil {
		return nil, err
	}
	if header != "SPLICE" {
		return nil, fmt.Errorf("Invalid file header, expected SPLICE, got %s", header)
	}

	size, err := readContentSize(f)
	if err != nil {
		return nil, err
	}

	version, err := readVersion(f)
	if err != nil {
		return nil, err
	}
	p.Version = version
	size -= 32

	tempo, err := readTempo(f)
	if err != nil {
		return nil, err
	}
	p.Tempo = tempo
	size -= 4

	var tracks []*Track
	for size > 0 {
		track, err := readTrack(f, &size)

		if err != nil {
			return nil, err
		}

		tracks = append(tracks, track)
	}
	p.Tracks = tracks

	return p, nil
}
