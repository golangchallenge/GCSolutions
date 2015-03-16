package drum

import (
	"fmt"
	"strconv"
)

// Pattern contained in a .splice file.
type Pattern struct {
	Version []byte
	Tempo   Tempo
	Tracks  []Track
}

// Tempo is the song's tempo, with nice formatting
type Tempo float32

// Track is a single instrument's beat pattern
type Track struct {
	ID    int32
	Name  []byte
	Steps Steps
}

// Steps of instrument's beat pattern
type Steps [16]byte

// Visual symbols for beat steps
var stepc = map[byte]byte{
	0: '-',
	1: 'x',
}

func (p Pattern) String() string {
	s := fmt.Sprintf("Saved with HW Version: %s\n", p.Version)
	s += fmt.Sprintf("Tempo: %s\n", p.Tempo)
	for _, t := range p.Tracks {
		s += fmt.Sprintf("%s\n", t)
	}
	return s
}

func (t Tempo) String() string {
	return strconv.FormatFloat(float64(t), 'f', -1, 32)
}

func (t Track) String() string {
	return fmt.Sprintf("(%d) %s\t%s", t.ID, t.Name, t.Steps)
}

func (s Steps) String() string {
	var r [16]byte

	for i, b := range s {
		r[i] = stepc[b]
	}

	return fmt.Sprintf("|%s|%s|%s|%s|",
		r[0:4], r[4:8], r[8:12], r[12:16])
}
