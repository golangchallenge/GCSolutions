package drum

import (
	"strings"
)

// A PatternHeader refers to the boilerplate above the musical notes in a pattern file.
type PatternHeader struct {
	_       [14]byte // An unused "tag", usually beginning with "SPLICE".
	Version [32]byte
	Tempo   float32
}

// GetVersion returns a Version which has been parsed and trimmed for
// further consumption.
func (ph *PatternHeader) GetVersion() Version {
	return Version(strings.TrimRight(string(ph.Version[:]), "\x00"))
}

// GetTempo returns a Tempo.
func (ph *PatternHeader) GetTempo() Tempo {
	return Tempo(ph.Tempo)
}
