package drum

import "testing"

func TestNewPattern(t *testing.T) {
	p := NewPattern()
	if p.Header == nil {
		t.Errorf("NewPattern hasn't initialized pattern's Header")
	}

	if len(p.Tracks) != 0 {
		t.Errorf("NewPattern created a Pattern with %d tracks instead of 0", len(p.Tracks))
	}
}

func TestPattern_AddTrack(t *testing.T) {
	p := NewPattern()
	tr := &Track{}
	p.AddTrack(tr)

	if len(p.Tracks) != 1 {
		t.Errorf("AddTrack hasn't added track to pattern's tracks")
	}
}
