package drum

import (
	"bytes"
	"log"
)

// StepSequence describes an entire 4/4, 16-step sequence.
type StepSequence struct {
	Steps []byte
}

// String implements the Stringer interface and returns
// a string representation of the step sequence.
func (s StepSequence) String() string {
	buf := bytes.NewBufferString("")

	for i, step := range s.Steps {
		if i%4 == 0 {
			_, err := buf.WriteString("|")
			if err != nil {
				log.Printf("Error writing to buffer: %v", err)
			}
		}

		if step == byte(0) {
			_, err := buf.WriteString("-")
			if err != nil {
				log.Printf("Error writing to buffer: %v", err)
			}
		} else if step == byte(1) {
			_, err := buf.WriteString("x")
			if err != nil {
				log.Printf("Error writing to buffer: %v", err)
			}
		}
	}

	_, err := buf.WriteString("|")
	if err != nil {
		log.Printf("Error writing to buffer: %v", err)
	}

	return buf.String()
}
