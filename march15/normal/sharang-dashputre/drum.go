// Package drum is supposed to implement the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
package drum

import "io/ioutil"

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var p Pattern
	p.readPattern(b)
	return &p, nil
}
