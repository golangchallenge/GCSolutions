package drum

import (
	"fmt"
	"os"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	p := &Pattern{}
	file, err := os.Open(path)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	defer file.Close()

	p.Decode(file)

	return p, nil
}
