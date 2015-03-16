package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	p := &Pattern{}

	file, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	header := make([]byte, 14)

	err = binary.Read(file, binary.LittleEndian, &header)

	if err != nil {
		return nil, err
	}

	version := make([]byte, 32)
	err = binary.Read(file, binary.LittleEndian, &version)

	if err != nil {
		return nil, err
	}

	p.Version = string(version[:bytes.Index(version, []byte{0})])

	err = binary.Read(file, binary.LittleEndian, &p.Tempo)

	if err != nil {
		return nil, err
	}

	parseTracks(file, p)

	return p, nil
}

func parseTracks(file io.Reader, pattern *Pattern) {
	for {
		var t Track

		err := binary.Read(file, binary.LittleEndian, &t.ID)
		if err == io.ErrUnexpectedEOF || err == io.EOF {
			break
		} else if err != nil {
			log.Panicln(err)
		}

		var unknown [3]byte
		err = binary.Read(file, binary.LittleEndian, &unknown)
		if err != nil {
			log.Panicln(err)
		}

		if t.ID == 83 && unknown == [3]byte{80, 76, 73} {
			// prolly that SPLICE segment again
			break
		}

		var instrumentLength uint8
		err = binary.Read(file, binary.LittleEndian, &instrumentLength)
		if err != nil {
			log.Panicln(err)
		}

		instrument := make([]byte, instrumentLength)

		err = binary.Read(file, binary.LittleEndian, &instrument)
		if err != nil {
			log.Panicln(err)
		}

		t.Name = string(instrument)

		t.Notes = make([]byte, 16)

		err = binary.Read(file, binary.LittleEndian, &t.Notes)

		if err != nil {
			log.Panicln(err)
		}

		pattern.Tracks = append(pattern.Tracks, t)
	}

}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version string
	Tempo   float32
	Tracks  []Track
}

func (p *Pattern) String() string {
	output := fmt.Sprintf("Saved with HW Version: %s\n", p.Version)
	output = output + fmt.Sprintf("Tempo: %g\n", p.Tempo)
	for _, t := range p.Tracks {
		output = output + t.String() + "\n"
	}

	return output
}

// Track is the content of each track
type Track struct {
	ID    uint8
	Name  string
	Notes []byte
}

func (t *Track) String() string {
	output := fmt.Sprintf("(%d) %s\t", t.ID, t.Name)
	for index, note := range t.Notes {
		if index%4 == 0 {
			output = output + "|"
		}
		if note == 0 {
			output = output + "-"
		} else {
			output = output + "x"
		}
	}
	return output + "|"
}
