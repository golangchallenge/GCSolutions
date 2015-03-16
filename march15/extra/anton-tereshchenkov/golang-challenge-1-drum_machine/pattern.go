package drum

import (
	"fmt"
	"math"
	"strings"
)

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version string  `splice:"32"`
	Tempo   float32 `splice:""` // "" is optional, can be omitted
	Tracks  []Track `splice:",end"`
}

// NewPattern creates a new Pattern.
func NewPattern(version string, tempo float32, tracks ...Track) *Pattern {
	return &Pattern{
		Version: version,
		Tempo:   tempo,
		Tracks:  tracks,
	}
}

// FindTrack searches for a track with id. Error is returned if Track is not found.
func (p *Pattern) FindTrack(id int) (*Track, error) {
	for _, t := range p.Tracks {
		if int(t.Id) == id {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("track %d not found", id)
}

// String implements fmt.Stringer interface. Used for printing.
func (p *Pattern) String() string {
	var lines []string
	lines = append(lines, fmt.Sprintf("Saved with HW Version: %s", p.Version))
	lines = append(lines, fmt.Sprintf("Tempo: %g", p.Tempo))
	for _, track := range p.Tracks {
		lines = append(lines, fmt.Sprint(track))
	}
	return strings.Join(lines, "\n") + "\n"
}

// Track is the representation of a drum track.
type Track struct {
	Id    int32  `splice:""`
	Name  string `splice:""`
	Steps []Step `splice:"16"`
}

// NewTrack creates a new Track with id and name. It also initializes
// 16 Steps (4/4 time signature, each step is a sixteenth note).
func NewTrack(id int, name string) *Track {
	return &Track{
		Id:    int32(id),
		Name:  name,
		Steps: make([]Step, 16),
	}
}

// Play sets which steps are played.
func (t *Track) Play(steps ...int) error {
	for _, s := range steps {
		if s >= len(t.Steps) {
			return fmt.Errorf("step index %d is out of bounds", s)
		}
		t.Steps[s].Play()
	}
	return nil
}

// Rest sets which steps are rests.
func (t *Track) Rest(steps ...int) error {
	for _, s := range steps {
		if s >= len(t.Steps) {
			return fmt.Errorf("step index %d is out of bounds", s)
		}
		t.Steps[s].Rest()
	}
	return nil
}

// partition splits a slice of steps into parts each with a maximum of n steps.
func (t *Track) partition(n int) [][]Step {
	// Get the number of parts
	s := float64(len(t.Steps)) / float64(n)
	parts := make([][]Step, int(math.Ceil(s)))
	for i, v := range t.Steps {
		parts[i/n] = append(parts[i/n], v)
	}
	return parts
}

// String implements fmt.Stringer interface. Used for printing.
func (t Track) String() string {
	steps := "|"
	// Partition steps into quarters
	for _, p := range t.partition(4) {
		for _, v := range p {
			steps += fmt.Sprintf("%v", v)
		}
		steps += "|"
	}
	return fmt.Sprintf("(%d) %s\t%s", t.Id, t.Name, steps)
}

// Step is a shortest available note. False step is a rest.
type Step bool

// Play indicates that the step is played.
func (n *Step) Play() {
	*n = Step(true)
}

// Rest indicates that the step is a rest.
func (n *Step) Rest() {
	*n = Step(false)
}

// String implements fmt.Stringer interface. Used for printing.
func (n Step) String() string {
	if n {
		return "x"
	}
	return "-"
}
