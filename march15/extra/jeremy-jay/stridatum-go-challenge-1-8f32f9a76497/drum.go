// Package drum is supposed to implement the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
package drum

import "fmt"

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	// HWVersion is the software that produced this Pattern.
	HWVersion string

	// Tempo contains the beats per minute for the Pattern.
	Tempo float32

	// Parts contains each of the instrumental parts that make up the Pattern.
	Parts []*Part
}

// Part is an individual instrument contained within a Pattern.
type Part struct {
	// ID is a unique ID representing this Part.
	ID int32

	// Name is a human-readable description of this Part.
	Name string

	// Steps is a bit-packed representation of the steps making up the Part,
	// with one bit per 16th note step. The high bit represents the beginning
	// of the measure and bit 0 represents the end of the measure.
	Steps uint16
}

var beats = make(map[int]string)

func init() {
	// build an encoded "beat" cache mapping to make printing faster/simpler
	for m := 0; m <= 0x0F; m++ {
		beat := []byte("----|")
		if (m & 8) == 8 {
			beat[0] = 'x'
		}
		if (m & 4) == 4 {
			beat[1] = 'x'
		}
		if (m & 2) == 2 {
			beat[2] = 'x'
		}
		if (m & 1) == 1 {
			beat[3] = 'x'
		}
		beats[m] = string(beat)
	}
}

func (p *Part) String() string {
	s := fmt.Sprintf("(%d) %s\t|", p.ID, p.Name)
	s += beats[int((p.Steps>>12)&0x0F)]
	s += beats[int((p.Steps>>8)&0x0F)]
	s += beats[int((p.Steps>>4)&0x0F)]
	s += beats[int(p.Steps&0x0F)]
	return s
}

func (p *Pattern) String() string {
	s := fmt.Sprintf("Saved with HW Version: %s\n", p.HWVersion)
	s += fmt.Sprintf("Tempo: %g\n", p.Tempo)
	for _, part := range p.Parts {
		s += part.String() + "\n"
	}
	return s
}
