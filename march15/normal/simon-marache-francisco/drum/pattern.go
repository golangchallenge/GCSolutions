package drum

import "fmt"

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	HWVersion string
	Tempo     float32
	Measures  []Measure
}

// Measure is the high level representation of a measure.
type Measure struct {
	Id    int32
	Name  string
	Steps [16]bool
}

// return the String representation of a Measure:
// (id) name  |xxxx|--xx|--xx|xxxx|
func (d Measure) String() (repr string) {
	repr = fmt.Sprintf("(%d) %s\t|", d.Id, d.Name)

	for i, b := range d.Steps {
		if i%4 == 0 && i > 0 {
			repr += "|"
		}
		if b {
			repr += "-"
		} else {
			repr += "x"
		}
	}
	repr += "|"
	return
}

// return the String representation of a Pattern
func (p Pattern) String() (repr string) {
	repr = "Saved with HW Version: " + p.HWVersion + "\n"
	repr += "Tempo: " + fmt.Sprintf("%g", p.Tempo) + "\n"
	for _, d := range p.Measures {
		repr += fmt.Sprint(d) + "\n"
	}
	return
}
