package drum

import (
	"bytes"
	"testing"
)

var TestID = []byte{0x01}

var TestName = []byte{
	0x00, 0x00, 0x00, 0x04, 0x6b, 0x69, 0x63, 0x6b,
}

var TestStepSequence = []byte{
	0x01, 0x00, 0x00, 0x00,
	0x01, 0x01, 0x00, 0x00,
	0x01, 0x01, 0x01, 0x00,
	0x01, 0x01, 0x01, 0x01,
}

func TestReadTrackID(t *testing.T) {
	track := &Track{}
	expected := uint8(1)
	reader := bytes.NewReader(TestID)

	read, err := readTrackID(reader, track)
	if err != nil {
		t.Errorf("Error reading track Id %v", err)
	}

	if track.ID != expected {
		t.Errorf("Mismatched track Id.  Expected: %d.  Actual: %d", expected, track.ID)
	}

	if read != TrackIDSize {
		t.Errorf("Mismatched bytes read. Expected %d.  Actual: %d", TrackIDSize, read)
	}
}

func TestReadTrackName(t *testing.T) {
	track := &Track{}
	expected := "kick"
	reader := bytes.NewReader(TestName)

	read, err := readTrackName(reader, track)
	if err != nil {
		t.Errorf("Error reading track name %v", err)
	}

	if track.Name != expected {
		t.Errorf("Mismatched track Name.  Expected: %s.  Actual: %s", expected, track.Name)
	}

	if read != TrackNameSize+len(expected) {
		t.Errorf("Mismatched bytes read. Expected %d.  Actual: %d", TrackNameSize+len(expected), read)
	}
}

func TestReadStepSequence(t *testing.T) {
	track := &Track{}
	expected := TestStepSequence
	reader := bytes.NewReader(TestStepSequence)

	read, err := readTrackStepSequence(reader, track)
	if err != nil {
		t.Errorf("Unable to reat step sequence")
	}

	for i := range expected {
		if track.StepSequence.Steps[i] != expected[i] {
			t.Errorf("Mismatched track StepSequence steps.  Expected: %v.  Actual: %v", expected[i], track.StepSequence.Steps[i])
		}
	}

	if read != StepSequenceSize {
		t.Errorf("Mismatched bytes read. Expected %d.  Actual: %d", StepSequenceSize, read)
	}

}
