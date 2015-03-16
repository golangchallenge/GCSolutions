package drum

import "testing"

func TestTrack_SetStep(t *testing.T) {
	track := &Track{
		ID:         1,
		Instrument: "kick",
	}
	err := track.SetStep(0, true)

	if err != nil {
		t.Fatalf("something went wrong setting step value - %v", err)
	}

	value := track.Steps[0]
	if value != 1 {
		t.Fatalf("step 0 has not been enabled")
	}

	err = track.SetStep(20, true)
	if err.Error() != "invalid step position `20`" {
		t.Fatalf("it should return an error if step position is out of range")
	}
}

func TestTrack_String(t *testing.T) {
	tr := &Track{
		ID:         9,
		Instrument: "kick",
	}

	tr.SetStep(0, true)
	tr.SetStep(2, true)

	expected := "(9) kick\t|x-x-|----|----|----|"
	s := tr.String()
	if s != expected {
		t.Errorf("Track's String method returned an unexpected string.\nexpected:\n%s\ngot:\n%s", expected, s)
	}
}
