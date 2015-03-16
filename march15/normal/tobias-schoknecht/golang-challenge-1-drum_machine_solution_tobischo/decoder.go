package drum

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
// TODO: implement
func DecodeFile(path string) (*Pattern, error) {
	p := &Pattern{}
	var file *os.File

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	r := bufio.NewReader(file)

	var data []byte
	if data, err = ioutil.ReadAll(r); err != nil {
		return nil, err
	}
	if err := p.UnmarshalBinary(data); err != nil {
		return nil, err
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

func (p Pattern) String() string {
	res := fmt.Sprintf("Saved with HW Version: %s\nTempo: %g\n", p.Version, p.Tempo)

	for _, track := range p.Tracks {
		res = fmt.Sprintf("%s%s\n", res, track)
	}

	return res
}

// Track is the representation of
// an individual track to the drum Pattern
type Track struct {
	ID    int32
	Name  string
	Steps []byte
}

func (t Track) String() string {
	steps := "|"

	for index, step := range t.Steps {
		if step == 1 {
			steps += "x"
		} else {
			steps += "-"
		}

		if index%4 == 3 {
			steps += "|"
		}
	}

	return fmt.Sprintf("(%d) %s\t%s", t.ID, t.Name, steps)
}
