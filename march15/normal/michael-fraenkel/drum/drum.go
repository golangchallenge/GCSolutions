// Package drum is supposed to implement the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
package drum

import (
	"fmt"
	"strconv"
	"strings"
)

type Track struct {
	Id    uint32
	Name  string
	Steps []byte
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
// TODO: implement
type Pattern struct {
	Version string
	Tempo   float32
	Tracks  []Track
}

func (t Track) String() string {
	var steps string

	for i, r := range t.Steps {
		if i%4 == 0 {
			steps = steps + "|"
		}

		if r == 0 {
			steps = steps + "-"
		} else {
			steps = steps + "x"
		}
	}

	steps = steps + "|"

	return fmt.Sprintf("(%d) %s\t%s", t.Id, t.Name, steps)
}

func (p Pattern) String() string {
	out := []string{
		"Saved with HW Version: " + p.Version,
		"Tempo: " + strconv.FormatFloat(float64(p.Tempo), 'f', -1, 32),
	}

	for _, t := range p.Tracks {
		out = append(out, t.String())
	}

	return strings.Join(out, "\n") + "\n"
}
