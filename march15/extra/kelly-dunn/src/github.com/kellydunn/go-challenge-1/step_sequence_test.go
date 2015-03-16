package drum

import (
	"testing"
)

func TestStepSequenceString(t *testing.T) {
	s := &StepSequence{
		Steps: make([]byte, 16),
	}

	expected := "|----|----|----|----|"

	if s.String() != expected {
		t.Errorf("Unexpected String for step sequence. \nExpected: %s \nActual: %s", expected, s.String())
	}

	s.Steps[0] = byte(1)

	expected = "|x---|----|----|----|"

	if s.String() != expected {
		t.Errorf("Unexpected String for step sequence. \nExpected: %s \nActual: %s", expected, s.String())
	}

	s.Steps[4] = byte(1)
	s.Steps[6] = byte(1)

	expected = "|x---|x-x-|----|----|"

	if s.String() != expected {
		t.Errorf("Unexpected String for step sequence. \nExpected: %s \nActual: %s", expected, s.String())
	}
}
