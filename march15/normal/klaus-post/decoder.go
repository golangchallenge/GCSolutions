package drum

import (
	"io"
	"os"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
// If the file cannot be found, an instance of *os.PathError is returned.
func DecodeFile(path string) (*Pattern, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	// We do not check error state of file.Close()
	// If everything proceeded until this point we have valid data.
	defer file.Close()
	return DecodeReader(file)
}

// DecodeReader decodes the drum machine from the supplied reader
// and return a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeReader(r io.Reader) (*Pattern, error) {
	p := &Pattern{}
	err := p.Decode(r)
	return p, err
}
