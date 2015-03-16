package drum

import (
	"bytes"
	"encoding/binary"
	//"encoding/hex"
	"fmt"
	//"github.com/kr/pretty"
	"io"
	"io/ioutil"
	"strings"
)

const (
	header       = "SPLICE"
	headerLen    = len(header)
	versionLen   = 32
	stepsInTrack = 16
)

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version [versionLen]byte
	Tempo   float32
	Tracks  []Track
}

// Track represents a single
type Track struct {
	ID    uint32
	Name  []byte
	Steps [stepsInTrack]byte
}

func (p Pattern) String() string {
	var s []string
	s = append(s, fmt.Sprintf("Saved with HW Version: %s", p.Version[:bytes.IndexByte(p.Version[:], 0x00)]))
	s = append(s, fmt.Sprintf("Tempo: %v", p.Tempo))
	for _, t := range p.Tracks {
		s = append(s, fmt.Sprintf("%s", t))
	}

	return strings.Join(s, "\n") + "\n"
}

func (t Track) String() string {
	var s []string
	s = append(s, fmt.Sprintf("(%d) %s\t", t.ID, t.Name))

	for i, b := range t.Steps {
		if i%4 == 0 {
			s = append(s, "|")
		}

		if b == 0x00 {
			s = append(s, "-")
		} else {
			s = append(s, "x")
		}
	}
	s = append(s, "|")

	return strings.Join(s, "")

}

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	p := &Pattern{}

	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %s", path, err)
	}

	if string(content[:headerLen]) != header {
		return nil, fmt.Errorf("incorrect file format")
	}

	r := bytes.NewReader(content[headerLen:])

	var patternLen uint64
	err = binary.Read(r, binary.BigEndian, &patternLen)
	if err != nil {
		return nil, fmt.Errorf("couldnt parse pattern length: %s", err)
	}

	dataStart := headerLen + 8
	dataEnd := uint64(dataStart) + patternLen
	trimmed := content[dataStart:dataEnd]
	r = bytes.NewReader(trimmed)

	err = binary.Read(r, binary.LittleEndian, &p.Version)
	if err != nil {
		return nil, fmt.Errorf("couldnt parse version: %s", err)
	}

	err = binary.Read(r, binary.LittleEndian, &p.Tempo)
	if err != nil {
		return nil, fmt.Errorf("couldnt parse tempo: %s", err)
	}

	p.Tracks, err = decodeTracks(r)
	if err != nil {
		return nil, fmt.Errorf("couldnt decode tracks: %s", err)
	}

	return p, nil
}

func decodeTracks(r io.Reader) ([]Track, error) {
	tracks := []Track{}

	for {
		t := &Track{}

		err := binary.Read(r, binary.LittleEndian, &t.ID)
		if err != nil {
			if err == io.EOF {
				// Little bit naive here but we are assuming
				// that Tracks will be formatted correctly and
				// as such a correct EOF will occur only in the
				// first read after the final Track.
				break
			}
			return nil, err
		}

		var nameLen uint8
		err = binary.Read(r, binary.LittleEndian, &nameLen)
		if err != nil {
			return nil, err
		}

		t.Name = make([]byte, nameLen)
		err = binary.Read(r, binary.LittleEndian, &t.Name)
		if err != nil {
			return nil, err
		}

		err = binary.Read(r, binary.LittleEndian, &t.Steps)
		if err != nil {
			return nil, err
		}

		tracks = append(tracks, *t)
	}

	return tracks, nil
}
