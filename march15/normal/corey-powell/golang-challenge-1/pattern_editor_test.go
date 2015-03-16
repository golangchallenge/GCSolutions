package drum

import (
	//"fmt"
	"testing"
)

func TestPatternEditor(t *testing.T) {
	pat := &Pattern{}
	pe := NewPatternEditor(pat)
	// ensure that EditTrack does returns an error if given an out of range index
	if _, err := pe.EditTrack(-1); err != IndexOutOfRange {
		t.Error("expected an IndexOutOfRange error for negative index.")
	}
	if _, err := pe.EditTrack(200); err != IndexOutOfRange {
		t.Error("expected an IndexOutOfRange error for too far index.")
	}
	// Track testing
	te, err := pe.CreateTrack("kick")
	// ensure that the track was created sucessfully
	if err != nil {
		t.Errorf("an error occured while creating track %v.", err)
	}
	// ensure that the track was added to the Pattern
	if len(pat.Tracks) != 1 {
		t.Error("Track was not added to Pattern.")
	}
	// ensure that the track added is indeed the same one returned.
	if pat.Tracks[0] != te.Track {
		t.Error("first Pattern Track does not match the created.")
	}
	if err := te.Toggle(0); err != nil {
		t.Errorf("an error occured while toggling a step %v.", err)
	}
	if err := te.Toggle(-1); err != IndexOutOfRange {
		t.Errorf("expected a IndexOutOfRange error.")
	}
	if err := te.Toggle(16); err != IndexOutOfRange {
		t.Errorf("expected a IndexOutOfRange error.")
	}
	if err := te.Toggle(0); err != nil {
		t.Errorf("an error occured while toggling a step %v.", err)
	}
	te.Clear()
	te.FillEvery(2, 0, 1)
	te.FillEvery(0, 0, 0)
	te.RotateLeft(1)
	te.RotateRight(4)
}
