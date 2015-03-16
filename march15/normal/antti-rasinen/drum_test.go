package drum

import "testing"

func TestTrackFormatting(t *testing.T) {
	expected := "(2) HiHat	|x-x-|x-x-|x-x-|x-x-|"
	track := Track{
		Name: "HiHat",
		Id:   2,
		Steps: [16]byte{
			1, 0, 1, 0,
			1, 0, 1, 0,
			1, 0, 1, 0,
			1, 0, 1, 0},
	}
	actual := track.String()
	if actual != expected {
		t.Fatalf("Got %s\nExpected %s", actual, expected)
	}
}
