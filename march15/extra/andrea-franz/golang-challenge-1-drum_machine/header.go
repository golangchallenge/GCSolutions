package drum

import "fmt"

// Header contains some info about the Pattern
type Header struct {
	Filename  string
	Signature string
	Tempo     float32
	Version   string
	X         byte
}

func (h *Header) String() string {
	return fmt.Sprintf("Saved with HW Version: %s\nTempo: %g", h.Version, h.Tempo)
}
