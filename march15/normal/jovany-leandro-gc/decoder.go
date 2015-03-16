package drum

import (
	"io/ioutil"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	p := &Pattern{}

	if err := p.UnmarshalBinary(data); err != nil {
		return nil, err
	}

	return p, nil
}
