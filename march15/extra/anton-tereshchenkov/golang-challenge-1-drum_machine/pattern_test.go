package drum

import (
	"fmt"
	"reflect"
	"testing"
)

func TestPattern_FindTrack(t *testing.T) {
	pattern := NewPattern("0", 120, *NewTrack(10, "kick"), *NewTrack(3, "snare"))
	cases := []struct {
		in   int
		want *Track
	}{
		{10, &pattern.Tracks[0]}, {3, &pattern.Tracks[1]},
	}
	for _, c := range cases {
		got, err := pattern.FindTrack(c.in)
		if err != nil {
			t.Errorf("FindTrack(%d): want %v, got error: %v", c.want, err)
		}
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("FindTrack(%d) == %q, want %q", c.in, got, c.want)
		}
	}
}

func TestPattern_FindTrack_error(t *testing.T) {
	pattern := NewPattern("0", 120, *NewTrack(10, "kick"))
	if got, err := pattern.FindTrack(1); err == nil {
		t.Errorf("FindTrack(%d) == %v, want error", 1, got)
	}
}

func TestPattern_String(t *testing.T) {
	pattern := NewPattern("0.808-alpha", 120, *NewTrack(10, "kick"), *NewTrack(3, "snare"))
	got := fmt.Sprint(pattern)
	want := fmt.Sprintf("Saved with HW Version: 0.808-alpha\nTempo: 120\n%s\n%s\n",
		pattern.Tracks[0], pattern.Tracks[1])
	if got != want {
		t.Errorf("String() == %q, want %q", got, want)
	}
}

func TestTrack_PlayRest(t *testing.T) {
	track := NewTrack(10, "kick")
	if err := track.Play(0, 3); err != nil {
		t.Errorf("Play(): got error: %v", err)
	}
	want := make([]Step, 16)
	want[0], want[3] = Step(true), Step(true)
	if !reflect.DeepEqual(track.Steps, want) {
		t.Errorf("Play(0, 3) == %v, want %v", track.Steps, want)
	}

	if err := track.Rest(3); err != nil {
		t.Errorf("Play(): got error: %v", err)
	}
	want[3] = Step(false)
	if !reflect.DeepEqual(track.Steps, want) {
		t.Errorf("Rest(3) == %v, want %v", track.Steps, want)
	}
}

func TestTrack_PlayRest_error(t *testing.T) {
	track := NewTrack(10, "kick")
	if err := track.Play(16); err == nil {
		t.Errorf("Play(16): want error")
		return
	}
	if err := track.Rest(16); err == nil {
		t.Errorf("Rest(16): want error")
		return
	}
}

func TestTrack_partition(t *testing.T) {
	track := NewTrack(10, "kick")
	cases := []struct {
		in, parts int
	}{
		{3, 6},
		{4, 4},
		{8, 2},
	}
	for _, c := range cases {
		got := track.partition(c.in)
		if len(got) != c.parts {
			t.Errorf("partition(%d) returned %d parts, want %d", c.in, len(got), c.parts)
		}

		total := 0
		for _, v := range got {
			total += len(v)
		}
		if total != len(track.Steps) {
			t.Errorf("partition(%d) returned %d steps, want %d", c.in, total, len(track.Steps))
		}
	}
}

func TestTrack_String(t *testing.T) {
	track := NewTrack(10, "kick")
	track.Play(0, 4, 8, 12)
	got, want := fmt.Sprint(track), "(10) kick\t|x---|x---|x---|x---|"
	if got != want {
		t.Errorf("String() == %q, want %q", got, want)
	}
}

func TestStep_Play(t *testing.T) {
	got, want := Step(false), Step(true)
	got.Play()
	if got != want {
		t.Errorf("Play() changed step to %t, want %t", got, want)
	}
}

func TestStep_Rest(t *testing.T) {
	got, want := Step(true), Step(false)
	got.Rest()
	if got != want {
		t.Errorf("Rest() changed step to %t, want %t", got, want)
	}
}

func TestStep_String(t *testing.T) {
	cases := []struct {
		step Step
		want string
	}{
		{Step(true), "x"},
		{Step(false), "-"},
	}
	for _, c := range cases {
		got := c.step.String()
		if got != c.want {
			t.Errorf("String() == %q, want %q", got, c.want)
		}
	}
}
