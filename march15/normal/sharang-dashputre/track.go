package drum

import "fmt"

// track is the reresentation of an individual track
// in a Pattern
type track struct {
	id    int
	name  string
	steps [lSteps]bool
}

func (t *track) String() string {
	s := fmt.Sprintf("(%d) %s\t", t.id, t.name)
	for i, val := range t.steps {
		if i%4 == 0 {
			s += "|"
		}
		if val {
			s += "x"
		} else {
			s += "-"
		}
	}
	return s + "|"
}

func (t *track) readTrack(b []byte, pos int) int {
	t.id = int(b[pos])
	lName := int(b[pos+4])
	pos += lName + 5
	t.name = string(b[pos-lName : pos])
	pos += lSteps
	for i, j := pos-lSteps, 0; i < pos; i, j = i+1, j+1 {
		t.steps[j] = b[i] != '\x00' // Convert binary to boolean
	}
	return pos
}
