package drum

import (
	"fmt"
)

// Instrument is a high level representation of the
// instrument in drum pattern
type Instrument struct {
	ID    byte
	Name  string
	Steps [16]byte
}

// Outputs drum pattern instrument to string
func (i *Instrument) String() (result string) {
	result += fmt.Sprintf("(%d) %s\t", i.ID, i.Name)
	for i, step := range i.Steps {
		if i%4 == 0 {
			result += "|"
		}
		if step != 0 {
			result += "x"
			continue
		}
		result += "-"
	}
	result += "|\n"
	return
}
