package drum

import "testing"

func TestStepPrint(t *testing.T) {
	tData := []struct {
		steps  Steps
		output string
	}{
		{
			Steps{1, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0},
			"|x---|x---|x---|x---|",
		},
	}
	for i, tc := range tData {
		o := tc.steps.String()
		if tc.output != o {
			t.Fatalf("%d wasn't printed as expect.\nGot:\n%s\nExpected:\n%s",
				i, o, tc.output)
		}
	}
}
