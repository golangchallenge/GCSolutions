// Package drum is supposed to implement the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
package drum

import (
	"bytes"
	"fmt"
	"math"
)

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version string
	Tempo   float32
	Tracks  []*Track
}

func (p *Pattern) String() string {

	tempo := fmt.Sprintf("%.0f", p.Tempo)
	if _, frac := math.Modf(float64(p.Tempo)); frac != 0 {
		tempo = fmt.Sprintf("%.1f", p.Tempo)
	}

	buf := bytes.NewBufferString(fmt.Sprintf("Saved with HW Version: %s\nTempo: %s\n", p.Version, tempo))
	for _, track := range p.Tracks {
		buf.WriteString(track.String() + "\n")
	}
	return buf.String()
}

type Track struct {
	Id   uint16
	Name string
	Steps
}

func (t *Track) String() string {
	return fmt.Sprintf("(%d) %s\t%s", t.Id, t.Name, t.Steps)
}

type Steps []byte

func (steps Steps) String() string {
	bf := bytes.NewBufferString("|")
	base := 0

	for base < len(steps) {
		i := 0
		for i < 4 {
			index := base + i
			if index > len(steps)-1 {
				break
			}
			bf.WriteString(step(steps[index]))
			i++
		}

		bf.WriteString("|")
		base += 4
	}

	return bf.String()
}

func step(u byte) string {
	switch u {
	case 0:
		return "-"
	case 1:
		return "x"
	default:
		panic("a step should be \x00 or \x01")
	}
}
