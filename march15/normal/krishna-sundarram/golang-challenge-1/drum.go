// Package drum is supposed to implement the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
package drum

import (
	"bytes"
	"fmt"
	"strconv"
)

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version string
	Tempo   float32
	Tracks  []Track
}

func (p Pattern) String() string {
	version := fmt.Sprintf("Saved with HW Version: %s\n", p.Version)
	tempo := fmt.Sprintf("Tempo: %s\n", strconv.FormatFloat(float64(p.Tempo), 'f', -1, 32))
	var b bytes.Buffer
	for i := 0; i < len(p.Tracks); i++ {
		b.WriteString(p.Tracks[i].String())
	}
	return version + tempo + b.String()
}

// Track represents an audio track that comprises a pattern.
type Track struct {
	Name  string
	Beats []int // will contain 1's and 0's
	ID    int
}

func (t Track) String() string {
	var b bytes.Buffer
	for i, number := range t.Beats {
		if i%4 == 0 {
			b.WriteString("|")
		}
		if number == 1 {
			b.WriteString("x")
		} else {
			b.WriteString("-")
		}
	}
	b.WriteString("|")
	return fmt.Sprintf("(%d) %s\t%s\n", t.ID, t.Name, b.String())
}
