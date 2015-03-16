package drum

import "os"

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	file, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	p := &Pattern{}

	err = p.Decode(file)

	if err != nil {
		return nil, err
	}

	return p, nil
}
