package drum

import (
	"path"
	"strconv"
	"testing"
)

func TestPlayPattern(t *testing.T) {
	// Play patterns 1-3 for 10 seconds each (4 & 5 sound pretty bad!).
	duration := 10
	for i := 1; i <= 3; i++ {
		pattern, err := DecodeFile(path.Join("fixtures", "pattern_"+strconv.Itoa(i)+".splice"))
		if err != nil {
			t.Fatalf("could not decode pattern %d: %v\n", i, err)
		}
		t.Logf("Playing pattern %d for %d seconds... ", i, duration)
		err = Play(pattern, duration)
		if err != nil {
			t.Fatalf("could not play pattern %d: %v\n", i, err)
		}
		t.Logf("done\n")
	}
}
