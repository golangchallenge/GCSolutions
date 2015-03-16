package drum

import "fmt"

// InvalidStepPositionError is returned after trying to
// modifying a step with an invalid step position.
type InvalidStepPositionError struct {
	position int
}

func (e InvalidStepPositionError) Error() string {
	return fmt.Sprintf("invalid step position `%d`", e.position)
}

// Track container the track ID, Instrument name and Steps.
type Track struct {
	ID         int32
	Instrument string
	Steps      [16]byte
}

// SetStep turns on/off a step at position `position`.
func (t *Track) SetStep(position int, on bool) error {
	if position > 15 {
		return InvalidStepPositionError{position: position}
	}

	if on {
		t.Steps[position] = 1
	} else {
		t.Steps[position] = 0
	}

	return nil
}

func (t *Track) String() string {
	s := fmt.Sprintf("(%d) %s\t|", t.ID, t.Instrument)
	for i := 0; i < 16; i++ {
		c := "-"
		if t.Steps[i] == 1 {
			c = "x"
		}
		s = fmt.Sprintf("%s%s", s, c)

		if (i+1)%4 == 0 {
			s = fmt.Sprintf("%s|", s)
		}
	}

	return s
}
