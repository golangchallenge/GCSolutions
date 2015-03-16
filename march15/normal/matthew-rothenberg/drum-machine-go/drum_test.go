package drum

import "testing"

/*
Most of the formatting is already tested quite well in the provided acceptance
tests, so just provide some simple validation of my helper functions as I build
them via TDD.
*/
func TestTrackString(t *testing.T) {
	track := &Track{
		ID:   5,
		Name: "cowbell",
		Beats: [16]Beat{
			true, true, true, true,
			false, false, false, false,
			true, false, true, false,
			false, true, false, true},
	}
	expect := "(5) cowbell\t|xxxx|----|x-x-|-x-x|\n"
	actual := track.String()

	if expect != actual {
		t.Fatalf("\nexpect: %s\nactual: %s", expect, actual)
	}
}

func TestBeatString(t *testing.T) {
	actual := Beat(false).String()
	expect := "-"
	if actual != expect {
		t.Fatalf("expect: %s, actual: %s", expect, actual)
	}
	actual = Beat(true).String()
	expect = "x"
	if expect != actual {
		t.Fatalf("expect: %s, actual: %s", expect, actual)
	}
}
