// Package drum is supposed to implement the decoding of .splice drum machine files.
package drum

import "fmt"

func (p *Pattern) String() string {
	str := fmt.Sprintf("Saved with HW Version: %s\nTempo: %g\n", p.Version, p.Tempo)

	for _, t := range p.Tracks {
		tStr := fmt.Sprintf("(%d) %s\t", t.ID, t.Name)

		meas := getMeasureStr(t)

		tStr += meas
		tStr += "\n"

		str += tStr
	}

	return str
}

func getMeasureStr(t Track) string {
	var m string
	for i := 0; i < 16; i++ {
		if i%4 == 0 {
			m += "|"
		}

		if t.Steps[i] == 0 {
			m += "-"
		} else {
			m += "x"
		}
	}
	m += "|"

	return m
}
