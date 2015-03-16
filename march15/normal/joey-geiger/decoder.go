package drum

import "os"

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	h := Header{}
	data, err := h.parse(f)
	if err != nil {
		return nil, err
	}

	p := &Pattern{}
	err = p.parse(data)

	return p, err
}
