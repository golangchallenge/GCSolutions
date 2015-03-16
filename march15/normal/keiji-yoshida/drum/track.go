package drum

import (
	"bytes"
	"strconv"
)

// maxLenShortName is the maximum length of
// a short name.
const maxLenShortName = 3

// Track represents a track of a drum machine file.
type Track struct {
	ID    int
	Name  []byte
	Steps []byte
}

// WriteTo writes the track data to the byte buffer.
func (t *Track) WriteTo(bf *bytes.Buffer) {
	bf.WriteString("(")
	bf.WriteString(strconv.Itoa(t.ID))
	bf.WriteString(") ")
	bf.Write(t.Name)
	if len(t.Name) <= maxLenShortName {
		bf.WriteString("\t\t")
	} else {
		bf.WriteString("\t")
	}
	for i, d := range t.Steps {
		if i%4 == 0 {
			bf.WriteString("|")
		}
		if d == 0x01 {
			bf.WriteString("x")
		} else {
			bf.WriteString("-")
		}
	}
	bf.WriteString("|")
}
