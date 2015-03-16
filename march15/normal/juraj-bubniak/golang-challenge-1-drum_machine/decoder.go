package drum

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
)

const (
	// stepOn is the value for enabled step.
	stepOn = "x"

	// stepOff  is the value for disabled step.
	stepOff = "-"

	// stepSep is the separator between steps.
	stepSep = "|"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	fp, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	p := &Pattern{}
	dec := NewDecoder(fp)
	err = dec.Decode(p)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version string
	Tempo   float32
	Tracks  []*track
}

// TempoString returns string representation of tempo, truncates decimals if possible.
func (p *Pattern) TempoString() string {
	intval := int(p.Tempo)
	trunc := float32(intval)

	if p.Tempo == trunc {
		return fmt.Sprintf("%d", intval)
	}

	return fmt.Sprintf("%.1f", p.Tempo)
}

// String returns string representation of pattern.
func (p *Pattern) String() string {
	var b bytes.Buffer

	b.WriteString(fmt.Sprintf("Saved with HW Version: %s\n", p.Version))
	b.WriteString(fmt.Sprintf("Tempo: %s\n", p.TempoString()))

	for _, t := range p.Tracks {
		b.WriteString(t.String())
	}

	return b.String()
}

// newTrack creates new track.
func newTrack(id int, name string, steps steps) *track {
	return &track{
		id:    id,
		name:  name,
		steps: steps,
	}
}

// track represents pattern track.
type track struct {
	id    int
	name  string
	steps steps
}

// String returns string representation of track.
func (t *track) String() string {
	var b bytes.Buffer
	b.WriteString("(")
	b.WriteString(strconv.Itoa(t.id))
	b.WriteString(") ")
	b.WriteString(t.name)
	b.WriteString("\t")
	b.WriteString(t.steps.String())
	b.WriteString("\n")
	return b.String()
}

// newStep creates new step.
func newStep(enabled bool) *step {
	return &step{enabled: enabled}
}

// step represents one of the 16 parts of track.
type step struct {
	enabled bool
}

// String returns string representation of step.
func (s *step) String() string {
	if s.enabled {
		return stepOn
	}
	return stepOff
}

// steps represents a sequence of steps.
type steps []*step

// String returns string representation of steps.
func (sps steps) String() string {
	var b bytes.Buffer

	for i, step := range sps {
		if i%4 == 0 {
			b.WriteString(stepSep)
		}
		b.WriteString(step.String())
	}

	b.WriteString(stepSep)
	return b.String()
}
