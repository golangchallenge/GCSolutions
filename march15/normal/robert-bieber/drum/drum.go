// Package drum is implements the decoding of .splice drum machine
// files.  See golang-challenge.com/go-challenge1/ for more
// information
package drum

import (
	"fmt"
	"strings"
)

// Active represents a sixteenth note in which a track is playing.
const Active Sixteenth = true

// Inactive represents a sixteenth note in which a track is not
// playing.
const Inactive Sixteenth = false

const sixteenthsPerMeasure = 16
const sixteenthsPerBeat = 4

// Sixteenth indicates whether a track is active or inactive for a
// given Sixteenth note.
type Sixteenth bool

// Measure represents a full measure's worth of sixteenth notes.
type Measure [sixteenthsPerMeasure]Sixteenth

// Track represents a single drum track.
type Track struct {
	ID   int32
	Name string
	Data Measure
}

// Pattern represents an entire splice file, including both its header
// data and its tracks.
type Pattern struct {
	FileName  string
	HWVersion string
	Tempo     float32
	Tracks    []*Track

	tracksLength uint32
}

func (p Pattern) String() string {
	lines := []string{
		fmt.Sprintf("Saved with HW Version: %s", p.HWVersion),
		fmt.Sprintf("Tempo: %g", p.Tempo),
	}

	for _, t := range p.Tracks {
		lines = append(lines, t.String())
	}

	return strings.Join(lines, "\n") + "\n"
}

func (t Track) String() string {
	return fmt.Sprintf("(%d) %s\t%v", t.ID, t.Name, t.Data)
}

func (m Measure) String() string {
	output := ""
	for k, v := range m {
		if k%sixteenthsPerBeat == 0 {
			output += "|"
		}
		output += v.String()
	}
	return output + "|"
}

func (s Sixteenth) String() string {
	if s == Active {
		return "x"
	}
	return "-"
}
