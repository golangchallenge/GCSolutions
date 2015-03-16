package drum

import (
	"fmt"
)

const tempoValue = 4

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	data, err := readFile(path)
	if err != nil {
		return nil, err
	}
	v, err := patterVersion(data)
	// error on header then all data cannot be a completed pattern
	if err != nil {
		return nil, err
	}
	t := patternTempo(data)
	ts := patternTracks(data)

	p := &Pattern{version: v, tempo: t, tracks: ts}
	return p, nil
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	version string
	tempo   string
	tracks  []Track
}

// Track struct to hold all information regarding a Track
type Track struct {
	id    string
	name  string
	steps []string
}

// Print Pattern
func (p *Pattern) String() string {
	result := ""
	for _, t := range p.tracks {
		res := fmt.Sprintf("(%s) %s\t", t.id, t.name)
		for i, val := range t.steps {
			if i%tempoValue == 0 {
				res = res + "|"
			}
			res = res + val
		}
		result = result + fmt.Sprintf("%s|\n", res)
	}
	return fmt.Sprintf("Saved with HW Version: %s\nTempo: %s\n%s", p.version, p.tempo, result)
}
