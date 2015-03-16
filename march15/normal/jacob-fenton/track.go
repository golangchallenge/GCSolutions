package drum

import "fmt"

// trackSteps is the number of steps within a track.
const trackSteps = 16

// A Track represents a single track used within
// a drum pattern.
type Track struct {
	ID   uint32
	Name string

	// Steps is a boolean array with each index
	// representing one step in the track, where
	// `true` means trigger the sound for a given
	// step, and `false` means don't.
	Steps [trackSteps]bool
}

func (t *Track) String() string {
	str := fmt.Sprintf("(%d) %s\t|", t.ID, t.Name)

	for i, trigger := range t.Steps {
		if trigger {
			str += "x"
		} else {
			str += "-"
		}

		// Insert separator every 4 steps.
		if (i+1)%4 == 0 {
			str += "|"
		}
	}

	return str + "\n"
}
