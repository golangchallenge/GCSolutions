package drum

import (
	"testing"
)

func TestPatternPrinting(t *testing.T) {
	p := NewPattern(eightZeroEightAlpha,
		120.4,
		map[int]Track{
			1: Track{
				"kick",
				[16]bool{
					false, false, false, true,
					true, false, false, false,
					true, false, true, false,
					true, false, false, false,
				},
			},
		})
	exp := `Saved with HW Version: 0.808-alpha
Tempo: 120.4
(1) kick	|---x|x---|x-x-|x---|
`

	if p.String() != exp {
		t.Logf("Pattern %#v did not print correctly.", p)
		t.Fatalf("Got:\n%v\n\nExpected:\n%v\n", p.String(), exp)
	}
}
