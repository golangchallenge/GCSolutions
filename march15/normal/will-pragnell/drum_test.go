package drum

import "testing"

func TestPatternString(t *testing.T) {
	tData := []struct {
		pattern Pattern
		output  string
	}{
		{Pattern{
			version: "an-version",
			tempo:   123.4,
			tracks: []track{
				track{
					id:    1,
					name:  "kick",
					steps: []bool{true, false, false, false, true, false, false, false, true, false, false, false, true, false, false, false},
				},
			},
		},
			`Saved with HW Version: an-version
Tempo: 123.4
(1) kick	|x---|x---|x---|x---|
`,
		},
		{Pattern{
			version: "an.version2",
			tempo:   99,
			tracks: []track{
				track{
					id:    42,
					name:  "snare",
					steps: []bool{false, false, true, false, true, false, false, false, false, false, true, false, true, false, false, true},
				},
			},
		},
			`Saved with HW Version: an.version2
Tempo: 99
(42) snare	|--x-|x---|--x-|x--x|
`,
		},
	}

	for _, exp := range tData {
		result := exp.pattern.String()
		if result != exp.output {
			t.Fatalf("Got:\n%s\nExpected:\n%s", result, exp.output)
		}
	}
}
