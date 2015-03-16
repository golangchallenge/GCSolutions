package drum

import (
	"bytes"
	"fmt"
	"os"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	p := &Pattern{}

	// it is important that these function calls are in order because they
	// all presume that the reader starts where their relevent data starts

	err = decodeFormatString(file)
	if err != nil {
		return nil, err
	}

	endPos, err := decodeFileLength(file)
	if err != nil {
		return nil, err
	}

	p.HWVersion, err = decodeHWVersion(file)
	if err != nil {
		return nil, err
	}

	p.Tempo, err = decodeTempo(file)
	if err != nil {
		return nil, err
	}

	// read tracks
	for {
		currPos, err := file.Seek(0, 1) // get current offset into file
		if err != nil {
			return nil, err
		}
		if currPos == endPos {
			// stop adding tracks when we reach the end of the file
			break
		}

		newTrack, err := decodeTrack(file)
		if err != nil {
			return nil, err
		}

		p.Tracks = append(p.Tracks, newTrack)
	}

	return p, nil
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	HWVersion string
	Tempo     float32
	Tracks    []Track
}

// Track represents what one instrument is supposed to play
type Track struct {
	ID    uint32
	Name  string
	Beats [16]byte
}

func (p Pattern) String() string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf(
		"Saved with HW Version: %s\nTempo: %g\n", p.HWVersion, p.Tempo,
	))

	for _, t := range p.Tracks {
		buffer.WriteString(t.String())
		buffer.WriteByte('\n')
	}

	return buffer.String()
}

func (t Track) String() string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf(
		"(%d) %s\t", t.ID, t.Name,
	))

	// write the beats and measures
	for i := range t.Beats {
		if i%4 == 0 {
			buffer.WriteByte('|') // separate measures
		}
		if t.Beats[i] == 0 {
			buffer.WriteByte('-')
		} else {
			buffer.WriteByte('x')
		}
	}
	buffer.WriteByte('|') // end with a measure bar

	return buffer.String()
}
