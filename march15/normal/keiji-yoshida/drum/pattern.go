package drum

import (
	"bytes"
	"fmt"
)

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version []byte
	Tempo   float32
	Tracks  []*Track
}

// String returns a string value of the pattern.
func (p *Pattern) String() string {
	bf := new(bytes.Buffer)
	bf.WriteString("Saved with HW Version: ")
	bf.Write(p.Version)
	bf.WriteString("\nTempo: ")
	bf.WriteString(fmt.Sprintf("%g", p.Tempo))
	bf.WriteString("\n")
	for _, t := range p.Tracks {
		t.WriteTo(bf)
		bf.WriteString("\n")
	}
	return bf.String()
}
