package drum

import (
	"fmt"
)

func (pattern *Pattern) String() string {
	output := fmt.Sprintf("Saved with HW Version: %s\nTempo: %g\n", pattern.Version, pattern.Tempo)
	for _, track := range pattern.Tracks {
		output += fmt.Sprintf("%s\n", track)
	}

	return output
}

func (track *Track) String() string {
	steps := "|"
	for i, step := range track.Steps {
		if step {
			steps += "x"
		} else {
			steps += "-"
		}
		if i%4 == 3 {
			steps += "|"
		}
	}

	output := fmt.Sprintf("(%d) %s\t%s", track.ID, track.Name, steps)

	return output
}
