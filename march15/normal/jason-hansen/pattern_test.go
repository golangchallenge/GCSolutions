package drum

import (
	"fmt"
	"testing"
)

func TestPatternString(t *testing.T) {
	tData := []struct {
		pattern Pattern
		output  string
	}{
		{Pattern{version: "0.808-alpha", tempo: float32(120),
			tracks: []track{{
				id: 0, name: "kick",
				steps: [16]step{
					1, 0, 0, 0,
					1, 0, 0, 0,
					1, 0, 0, 0,
					1, 0, 0, 0,
				}}, {
				id: 1, name: "snare",
				steps: [16]step{
					0, 0, 0, 0,
					1, 0, 0, 0,
					0, 0, 0, 0,
					1, 0, 0, 0,
				}}, {
				id: 2, name: "clap",
				steps: [16]step{
					0, 0, 0, 0,
					1, 0, 1, 0,
					0, 0, 0, 0,
					0, 0, 0, 0,
				}}, {
				id: 3, name: "hh-open",
				steps: [16]step{
					0, 0, 1, 0,
					0, 0, 1, 0,
					1, 0, 1, 0,
					0, 0, 1, 0,
				}}, {
				id: 4, name: "hh-close",
				steps: [16]step{
					1, 0, 0, 0,
					1, 0, 0, 0,
					0, 0, 0, 0,
					1, 0, 0, 1,
				}}, {
				id: 5, name: "cowbell",
				steps: [16]step{
					0, 0, 0, 0,
					0, 0, 0, 0,
					0, 0, 1, 0,
					0, 0, 0, 0,
				}}}},
			`Saved with HW Version: 0.808-alpha
Tempo: 120
(0) kick	|x---|x---|x---|x---|
(1) snare	|----|x---|----|x---|
(2) clap	|----|x-x-|----|----|
(3) hh-open	|--x-|--x-|x-x-|--x-|
(4) hh-close	|x---|x---|----|x--x|
(5) cowbell	|----|----|--x-|----|
`}, {Pattern{version: "0.708-alpha", tempo: float32(999),
			tracks: []track{{
				id: 1, name: "Kick",
				steps: [16]step{
					1, 0, 0, 0,
					0, 0, 0, 0,
					1, 0, 0, 0,
					0, 0, 0, 0,
				}}, {
				id: 2, name: "HiHat",
				steps: [16]step{
					1, 0, 1, 0,
					1, 0, 1, 0,
					1, 0, 1, 0,
					1, 0, 1, 0,
				}}}},
			`Saved with HW Version: 0.708-alpha
Tempo: 999
(1) Kick	|x---|----|x---|----|
(2) HiHat	|x-x-|x-x-|x-x-|x-x-|
`}}

	for _, d := range tData {
		if fmt.Sprint(d.pattern) != d.output {
			t.Logf("decoded:\n%#v\n", fmt.Sprint(d.pattern))
			t.Logf("expected:\n%#v\n", d.output)
			t.Fatalf("pattern did not output to string correctly.\nGot:\n%s\nExpected:\n%s",
				d.pattern, d.output)
		}
	}
}
