package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
)

const spliceExt string = "SPLICE"

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	p := &Pattern{}

	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	r := bytes.NewReader(b)

	var ext [6]byte
	if err := binary.Read(r, binary.LittleEndian, &ext); err != nil {
		return nil, err
	}
	if string(ext[:]) != spliceExt {
		return nil, fmt.Errorf("drum: decoder received unknown file %g", string(ext[:]))
	}

	var dataSize int64
	if err := binary.Read(r, binary.BigEndian, &dataSize); err != nil {
		return nil, err
	}

	b = b[14 : 14+dataSize]
	r = bytes.NewReader(b)
	err = p.Decode(r)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version string
	Tempo   float32
	Tracks  []*Track
}

// Track is a part of a drum pattern that
// represents a grid of 16 parts to trigger a sound.
type Track struct {
	Id    int32
	Name  string
	Steps [16]uint8
}

func (p *Pattern) Decode(r io.Reader) error {
	var v [32]byte
	if err := binary.Read(r, binary.LittleEndian, &v); err != nil {
		return err
	}
	p.Version = strings.TrimRight(string(v[:]), "\x00")

	if err := binary.Read(r, binary.LittleEndian, &p.Tempo); err != nil {
		return err
	}

	// Loop to read all the tracks available till the
	// end of file is reached.
	for {
		t := &Track{}
		if err := binary.Read(r, binary.LittleEndian, &t.Id); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		var l byte
		if err := binary.Read(r, binary.LittleEndian, &l); err != nil {
			return err
		}

		n := make([]byte, int(l))
		if err := binary.Read(r, binary.LittleEndian, &n); err != nil {
			return err
		}
		t.Name = string(n[:])

		if err := binary.Read(r, binary.LittleEndian, &t.Steps); err != nil {
			return err
		}

		p.Tracks = append(p.Tracks, t)
	}

	return nil
}

func (p *Pattern) String() string {
	var t bytes.Buffer

	for _, i := range p.Tracks {
		t.WriteString(i.String())
	}

	return fmt.Sprintf("Saved with HW Version: %s\nTempo: %g\n%s", p.Version, p.Tempo, t.String())
}

func (t *Track) String() string {
	var s bytes.Buffer

	for n, i := range t.Steps {
		if n%4 == 0 && n != 0 {
			s.WriteString("|")
		}
		if i == 0 {
			s.WriteString("-")
			continue
		}
		s.WriteString("x")
	}

	return fmt.Sprintf("(%d) %s\t|%s|\n", t.Id, t.Name, s.String())
}
