package drum

import (
	"io/ioutil"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	// assuming that file is not very big
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	p := &Pattern{}
	err = decode(data, p)
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
	Tracks  []Track
}

// A Track represents a single track inside a Pattern
type Track struct {
	ID    uint8
	Name  string
	Steps [16]bool
}
