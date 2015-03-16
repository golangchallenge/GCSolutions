package drum

import (
	"strings"
	"testing"
)

func TestPatternString(t *testing.T) {
	p := Pattern{
		Version: []byte("alphabeta"),
		Tempo:   Tempo(123.15),
		Tracks:  []Track{Track{}, Track{}},
	}
	x := p.String()
	s := `Saved with HW Version: alphabeta
Tempo: 123.15
`
	if !strings.Contains(x, s) {
		t.Fatalf("Pattern string received:\n%s\ndoes not contain:\n%s",
			x, s)
	}
	if len(strings.Split(x, "\n")) != 5 {
		t.Fatalf("Pattern string received:\n%s\nis not of 5 lines", x)
	}
}

func TestTempoString(t *testing.T) {
	exs := map[float32]string{
		50:     "50",
		50.5:   "50.5",
		50.005: "50.005",
	}
	for k, v := range exs {
		m := Tempo(k)
		s := m.String()
		if v != s {
			t.Fatalf("Wrong tempo string %s, expected %s", s, v)
		}
	}
}

func TestStepsString(t *testing.T) {
	st := Steps{
		0, 0, 0, 1,
		0, 1, 0, 1,
		1, 0, 1, 1,
		1, 0, 0, 0,
	}.String()
	s := "|---x|-x-x|x-xx|x---|"
	if st != s {
		t.Fatalf("Wrong steps string %s, expected %s", s, st)
	}
}

func TestTrackString(t *testing.T) {
	ks := Track{
		ID:   15,
		Name: []byte("Didgeridoo"),
	}.String()
	s := "(15) Didgeridoo\t"
	if !strings.Contains(ks, s) {
		t.Fatalf("Track string %s, does not contain expected %s", ks, s)
	}
}
