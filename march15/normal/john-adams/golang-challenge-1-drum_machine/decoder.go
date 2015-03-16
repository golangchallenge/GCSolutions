package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	p := &Pattern{}

	s, err := ioutil.ReadFile(path)
	buf := bytes.NewReader(s)

	// Let's process the header first, this will tell us the Type of file (SLICE)
	// as well as the length of the Pattern
	err = binary.Read(buf, binary.LittleEndian, &p.Header)

	// Let's calculate the EOF from the PatternLength.  This will let us get
	// around the eroneous data in pattern_5
	eof := buf.Len() - int(p.Header.PatternLength)

	// Read out the HW Version
	binary.Read(buf, binary.LittleEndian, &p.Version)

	// Read out the Tempo
	binary.Read(buf, binary.LittleEndian, &p.Tempo)

	// The remaining data should be tracks, so let's process them
	for i := eof; i < buf.Len(); {
		t := Track{}

		// Read out the ID of the track
		binary.Read(buf, binary.LittleEndian, &t.ID)

		// Read out the length of the name
		binary.Read(buf, binary.LittleEndian, &t.NameLength)

		// Set the size of the Name of the track
		t.Name = make([]byte, t.NameLength)

		// Read out the name
		binary.Read(buf, binary.LittleEndian, &t.Name)

		// Read out the steps
		binary.Read(buf, binary.LittleEndian, &t.Steps)

		p.Tracks = append(p.Tracks, t)
	}
	return p, err
}

// A Header represents the Header for a drum Pattern.
type Header struct {
	Type          [13]byte
	PatternLength uint8
}

// A Track represetnts an instance of a Track for a drum Pattern.
type Track struct {
	ID         int32
	NameLength uint8
	Name       []byte
	Steps      [16]byte
}

// String outputs a formatted string of the Track
func (t *Track) String() string {
	value := fmt.Sprintf("(%v) %s\t|", t.ID, t.Name)

	for i, step := range t.Steps {
		if step == 0x01 {
			value += "x"
		} else {
			value += "-"
		}
		if i > 0 && (i+1)%4 == 0 {
			value += "|"
		}
	}
	value += "\n"
	return value

}

// A Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Header  Header
	Version [32]byte
	Tempo   float32
	Tracks  []Track
}

// String outputs a formatted string of the Pattern
func (p *Pattern) String() string {
	value := fmt.Sprintf("Saved with HW Version: %s\nTempo: %v\n", bytes.Trim(p.Version[:], "\x00"), p.Tempo)
	for _, track := range p.Tracks {
		value += track.String()
	}
	return value
}
