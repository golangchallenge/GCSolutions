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
	Version []byte
	Tempo   float32
	Tracks  []Track
}

// Track is struct to contain track information
type Track struct {
	Id    uint8
	Name  []byte
	Steps [16]byte
}

func (t *Track) String() string {
	steps := make([]byte, 16)
	for idx, b := range t.Steps {
		if b == 0x01 {
			steps[idx] = 0x78
		} else {
			steps[idx] = 0x2d
		}
	}
	return fmt.Sprintf("(%d) %s\t|%s|%s|%s|%s|\n", t.Id, t.Name, steps[:4], steps[4:8], steps[8:12], steps[12:])
}

func (p *Pattern) String() string {
	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf("Saved with HW Version: %s\nTempo: %s\n", p.Version, strconv.FormatFloat(float64(p.Tempo), 'f', -1, 32)))

	for _, t := range p.Tracks {
		buffer.WriteString(fmt.Sprint(&t))
	}

	return buffer.String()
}
