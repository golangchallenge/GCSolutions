// Package drum is supposed to implement the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
package drum

import (
	"bytes"
	"fmt"
	"io"
)

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version string
	Tempo   float32
	Tracks  []Track
}

func (p *Pattern) String() string {
	var buf bytes.Buffer

	// truncate the version if it's too long
	ver := p.Version
	if len(ver) > 32 {
		ver = ver[:32]
	}

	io.WriteString(&buf, "Saved with HW Version: "+ver+"\n")
	io.WriteString(&buf, fmt.Sprintf("Tempo: %g\n", p.Tempo))
	for _, track := range p.Tracks {
		fmt.Fprintln(&buf, track.String())
	}
	return string(buf.Bytes())
}

// AddTrack adds a track to the pattern.
func (p *Pattern) AddTrack(t Track) {
	p.Tracks = append(p.Tracks, t)
}

// Track is an audio sample consisting of 16 measures.
// A single pattern can be made up of multiple tracks.
type Track struct {
	ID      uint32
	Name    string
	measure uint16
}

// StepActive returns true if the specified step (0 through 15)
// is active in the track's measure, or false otherwise.
func (t *Track) StepActive(step uint16) bool {
	return getBit(step, t.measure)
}

// ToggleStep either enables or disables the sound at the specified
// step in the track's measure.
func (t *Track) ToggleStep(step uint16, enabled bool) {
	if enabled {
		t.measure |= 1 << step
	} else {
		t.measure &= ^uint16(1 << step)
	}
}

func (t *Track) String() string {
	return fmt.Sprintf("(%d) %s\t%s", t.ID, t.Name, t.measureString())
}

func (t *Track) measureString() string {
	var b bytes.Buffer
	b.WriteByte('|')
	for i := uint16(0); i < 16; i++ {
		if t.StepActive(i) {
			b.WriteByte('x')
		} else {
			b.WriteByte('-')
		}

		if (i+1)%4 == 0 {
			b.WriteByte('|')
		}
	}
	return string(b.Bytes())
}

func getBit(bit, value uint16) bool {
	if bit > 15 {
		return false
	}
	return value&(1<<bit) > 0
}
