package drum

import (
	"fmt"
)

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	version     string
	Tempo       float32
	Instruments []*Instrument
	footer      []byte
}

// Outputs drum pattern to string
func (p *Pattern) String() (result string) {
	result += fmt.Sprintf("Saved with HW Version: %s\n", p.version)
	if float32(int32(p.Tempo)) != p.Tempo {
		result += fmt.Sprintf("Tempo: %.1f\n", p.Tempo)
	} else {
		result += fmt.Sprintf("Tempo: %.0f\n", p.Tempo)
	}
	for _, instrument := range p.Instruments {
		result += fmt.Sprint(instrument)
	}
	return
}
