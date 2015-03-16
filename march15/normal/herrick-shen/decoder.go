package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
// TODO: implement
func DecodeFile(path string) (*Pattern, error) {
	p := &Pattern{}

	// Read file to data
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return p, err
	}

	// Create new reader
	r := bytes.NewReader(data)

	// const error
	errInvalid := errors.New("data invalid")

	// If data not complete
	if r.Len() <= 14 {
		return p, errInvalid
	}

	// Helper function, trim bytes to string
	trimString := func(buf []byte) string {
		return strings.TrimRight(string(buf[:]), string([]byte{0}))
	}

	// Check first part is "SPLICE"
	SPLICE := make([]byte, 10)
	binary.Read(r, binary.BigEndian, &SPLICE)
	if trimString(SPLICE) != "SPLICE" {
		return p, errInvalid
	}

	// Read file len
	var fileLen uint32
	binary.Read(r, binary.BigEndian, &fileLen)
	if uint32(r.Len()) < fileLen {
		return p, errInvalid
	}

	// Read version
	version := make([]byte, 32)
	binary.Read(r, binary.LittleEndian, &version)
	p.Version = trimString(version)

	// Read tempo
	binary.Read(r, binary.LittleEndian, &p.Tempo)

	// Read tracks
	p.Tracks = make([]Track, 0)
	if r.Len() == 0 {
		return p, nil
	}

	for {
		track := Track{}

		// Read track id
		if err = binary.Read(r, binary.BigEndian, &track.Id); err != nil {
			break
		}

		// Read track name len
		var nameLen uint32
		if err = binary.Read(r, binary.BigEndian, &nameLen); err != nil {
			break
		}

		// Read track name
		trackName := make([]byte, nameLen)
		if err = binary.Read(r, binary.BigEndian, trackName); err != nil {
			break
		}
		track.Name = trimString(trackName)

		// Read steps
		if err = binary.Read(r, binary.BigEndian, &track.Steps); err != nil {
			break
		}

		// Append track
		p.Tracks = append(p.Tracks, track)

		// If read finished, then break the loop
		if r.Len() == 0 {
			break
		}
	}

	return p, nil
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
// TODO: implement
type Pattern struct {
	Version string
	Tempo   float32
	Tracks  []Track
}

type Track struct {
	Id    byte
	Name  string
	Steps [16]byte
}

func (p *Pattern) String() string {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("Saved with HW Version: %s\n", p.Version))
	buf.WriteString(fmt.Sprintf("Tempo: %v\n", p.Tempo))

	for _, t := range p.Tracks {
		buf.WriteString(fmt.Sprintf("(%d) %s\t", t.Id, t.Name))

		for i := 0; i < 16; i++ {
			if i%4 == 0 {
				buf.WriteString("|")
			}

			if t.Steps[i] > 0 {
				buf.WriteString("x")
			} else {
				buf.WriteString("-")
			}
		}
		buf.WriteString("|\n")
	}

	return string(buf.Bytes())
}
