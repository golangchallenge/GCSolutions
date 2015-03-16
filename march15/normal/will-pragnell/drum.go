package drum

import "fmt"

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	version string
	tempo   float32
	tracks  []track
}

func (p Pattern) String() string {
	result := fmt.Sprintf("Saved with HW Version: %v\nTempo: %v\n", p.version, p.tempo)
	for _, track := range p.tracks {
		result += fmt.Sprint(track)
	}
	return result
}

// track is the representation of each track within a pattern
type track struct {
	id    uint8
	name  string
	steps []bool
}

func (t track) String() string {
	result := fmt.Sprintf("(%v) %v	|", t.id, t.name)
	for i, step := range t.steps {
		if step {
			result += "x"
		} else {
			result += "-"
		}

		if i%4 == 3 {
			result += "|"
		}
	}
	return result + "\n"
}
