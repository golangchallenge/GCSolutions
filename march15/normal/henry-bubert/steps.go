package drum

import "fmt"

// Track represents an ID, Name and the pattern for a single track
type Track struct {
	ID    int
	Name  string
	Steps Steps
}

// prints a formatted track with ID, Name and it's setps
func (t Track) String() string {
	return fmt.Sprintf("(%d) %s\t%s\n", t.ID, t.Name, t.Steps)
}

// how many steps per track
const stepCnt = 16

// Steps holds stepCnt bools (one step is a bool) and can print them as ascii
type Steps [stepCnt]byte

// String returns a pattern grouped into 4 steps sorrounded by |. 1 == x, 0 == -
func (s Steps) String() string {
	var o string
	for i := 0; i < stepCnt; i++ {
		if i%4 == 0 {
			o += "|"
		}
		if s[i] == 1 {
			o += "x"
		} else {
			o += "-"
		}
	}
	o += "|"
	return o
}
