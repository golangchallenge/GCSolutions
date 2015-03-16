package drum

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
)

// ErrBadSignature means that an invalid signature was read
// when parsing a splice.
var ErrBadSignature = errors.New("invalid signature")

// magicHeader represents the expected header in splice files.
// Failing to read this header means the file is not a splice.
const magicHeader = "SPLICE\x00\x00\x00\x00\x00\x00\x00"

// Measure represents a grid of 16 steps, each
// of which can be either on (1) or off (0).
type Measure [16]byte

// Track is the high level representation of
// a drum track divided into 16 steps.
type Track struct {
	ID    uint32
	Name  string
	Steps Measure
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	HWVersion string
	Tempo     float32
	Tracks    []Track
}

// String returns a pretty textual representation of the pattern.
func (p *Pattern) String() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "Saved with HW Version: %s\n", p.HWVersion)
	fmt.Fprintf(&buf, "Tempo: %s\n", strconv.FormatFloat(float64(p.Tempo), 'f', -1, 32))

	for _, track := range p.Tracks {
		fmt.Fprintf(&buf, "(%d) %s\t", track.ID, track.Name)
		for i, value := range track.Steps {
			// print bar
			if i%4 == 0 {
				buf.WriteByte('|')
			}

			// print value
			if value != 0 {
				buf.WriteByte('x')
			} else {
				buf.WriteByte('-')
			}
		}
		buf.WriteString("|\n")
	}

	return buf.String()
}
