package drum

import (
	"bytes"
	"fmt"
)

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	version string
	tempo   float32
	tracks  []*Track
}

// SetVersion assigns the version string to the pattern instance.
func (p *Pattern) SetVersion(v string) {
	p.version = v
}

// Version returns the patterns version number.
func (p *Pattern) Version() string {
	return p.version
}

// SetTempo assigns the tempo for the pattern instance. This is in bpm (beats per minute).
func (p *Pattern) SetTempo(t float32) {
	p.tempo = t
}

// Tempo returns the tempo of the pattern. This is in bpm (beats per minute).
func (p *Pattern) Tempo() float32 {
	return p.tempo
}

// AddTrack will append the given track instance to the pattern.
func (p *Pattern) AddTrack(b *Track) {
	p.tracks = append(p.tracks, b)
}

// String will return a string value of the pattern; this includes version, tempo and the
// tracks and there associated steps in 4/4 format.
func (p *Pattern) String() string {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("Saved with HW Version: %v\nTempo: %v\n", p.version, p.tempo))
	for _, b := range p.tracks {
		buf.WriteString(fmt.Sprintln(b.String()))
	}

	return buf.String()
}

// NewTrack creates a new instance of a track with fixed number of steps.
func NewTrack() *Track {
	return &Track{
		steps: make([]bool, 16),
	}
}

// Track represents the 4/4 set of steps for a name/id.
type Track struct {
	id    uint32
	name  string
	steps []bool
}

// ID returns the tracks integer identifier.
func (t *Track) ID() uint32 {
	return t.id
}

// SetID sets the identifier number of the track.
func (t *Track) SetID(id uint32) {
	t.id = id
}

// Name is the tracks user-friendly representation.
func (t *Track) Name() string {
	return t.name
}

// SetName sets the name of the track.
func (t *Track) SetName(name string) {
	t.name = name
}

// SetStep assigns the given step index the value true/false.
func (t *Track) SetStep(n int, v bool) {
	t.steps[n] = v
}

// Step returns the n'th index value - whether it is true/false.
func (t *Track) Step(n int) bool {
	return t.steps[n]
}

// String returns the id, name and steps of the track as a string.
func (t *Track) String() string {

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("(%v) %v\t", t.ID(), t.Name()))

	for i, s := range t.steps {

		if i%4 == 0 {
			buf.WriteByte('|')
		}

		if s {
			buf.WriteByte('x')
		} else {
			buf.WriteByte('-')
		}
	}

	buf.WriteByte('|')
	return buf.String()
}
