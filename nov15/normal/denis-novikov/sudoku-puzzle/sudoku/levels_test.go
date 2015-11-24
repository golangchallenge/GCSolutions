package sudoku

import "testing"

func TestLevelToString(t *testing.T) {
	tests := []struct {
		level DifficultyLevel
		str   string
	}{
		{DLUnknown, "Unknown"},
		{DLEasy, "Easy"},
		{DLMedium, "Medium"},
		{DLHard, "Hard"},
	}

	for _, test := range tests {
		if test.level.String() != test.str {
			t.Error("Wrong string returned: ", test.level.String(),
				" instead of ", test.str)
		}
	}
}

func TestLevel(t *testing.T) {
	levels := [...]DifficultyLevel{DLEasy, DLMedium, DLHard}
	for i, level := range levels[1:] {
		if !(level > levels[i]) {
			t.Error("Level \"", level, "\" must be greater "+
				"than level \"", levels[i], "\n")
		}
	}
}
