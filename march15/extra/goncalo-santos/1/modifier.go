package drum

import "fmt"

// ReplaceSteps replaces instrument's pattern to
// what is passed in newPattern
func (p *Pattern) ReplaceSteps(instrumentID byte, newPattern [16]byte) error {
	if _, ok := p.instruments[instrumentID]; !ok {
		return fmt.Errorf("no instrument with ID=%d exists to be modified", instrumentID)
	}

	p.instruments[instrumentID].Pattern = newPattern

	return nil
}

// AddSteps adds steps to the instrument
// The 1s in stepsToAdd will be 1s in the new pattern; the 0s will stay what they were before
// Example:
// p.instruments[0].Pattern = {1, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0}
// p.AddSteps(0, [16]byte{0, 1, 1, 1, 0, 1, 1, 1, 0, 1, 1, 1, 0, 1, 1, 1})
// p.instruments[0].Pattern = {1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
func (p *Pattern) AddSteps(instrumentID byte, stepsToAdd [16]byte) error {
	if _, ok := p.instruments[instrumentID]; !ok {
		return fmt.Errorf("no instrument with ID=%d exists to be modified", instrumentID)
	}

	for i, val := range stepsToAdd {
		p.instruments[instrumentID].Pattern[i] = p.instruments[instrumentID].Pattern[i] | val
	}

	return nil
}

// RemoveSteps removes steps from the instrument
// The 1s in stepsToRemove will be 0s in the new pattern; the 0s will stay what they were before
// Example:
// p.instruments[0].Pattern = {1, 1, 0, 0, 1, 1, 0, 0, 1, 1, 0, 0, 1, 1, 0, 0}
// p.AddSteps(0, [16]byte{1, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0})
// p.instruments[0].Pattern = {0, 1, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0}
func (p *Pattern) RemoveSteps(instrumentID byte, stepsToRemove [16]byte) error {
	if _, ok := p.instruments[instrumentID]; !ok {
		return fmt.Errorf("no instrument with ID=%d exists to be modified", instrumentID)
	}

	for i, val := range stepsToRemove {
		p.instruments[instrumentID].Pattern[i] = p.instruments[instrumentID].Pattern[i] &^ val
	}

	return nil
}
