package drummachine

import (
	"reflect"
	"testing"

	"github.com/nf/sigourney/audio"

	"github.com/rogpeppe/misc/drum"
)

var trackTests = []struct {
	track        drum.Track
	beatDuration int64
	nextAfter    int64
	expect       int64
}{{
	track: drum.Track{
		Beats: [drum.NumBeats]bool{0: true},
	},
	beatDuration: 10,
	nextAfter:    0,
	expect:       0,
}, {
	track: drum.Track{
		Beats: [drum.NumBeats]bool{0: true},
	},
	beatDuration: 10,
	nextAfter:    1,
	expect:       160,
}, {
	track: drum.Track{
		Beats: [drum.NumBeats]bool{1: true},
	},
	beatDuration: 10,
	nextAfter:    0,
	expect:       10,
}, {
	track: drum.Track{
		Beats: [drum.NumBeats]bool{1: true},
	},
	beatDuration: 10,
	nextAfter:    11,
	expect:       170,
}, {
	track: drum.Track{
		Beats: [drum.NumBeats]bool{1: true, 5: true, 15: true},
	},
	beatDuration: 10,
	nextAfter:    0,
	expect:       10,
}, {
	track: drum.Track{
		Beats: [drum.NumBeats]bool{1: true, 5: true, 15: true},
	},
	beatDuration: 10,
	nextAfter:    9,
	expect:       10,
}, {
	track: drum.Track{
		Beats: [drum.NumBeats]bool{1: true, 5: true, 15: true},
	},
	beatDuration: 10,
	nextAfter:    10,
	expect:       10,
}, {
	track: drum.Track{
		Beats: [drum.NumBeats]bool{1: true, 5: true, 15: true},
	},
	beatDuration: 10,
	nextAfter:    11,
	expect:       50,
}, {
	track: drum.Track{
		Beats: [drum.NumBeats]bool{1: true, 5: true, 15: true},
	},
	beatDuration: 10,
	nextAfter:    50,
	expect:       50,
}, {
	track: drum.Track{
		Beats: [drum.NumBeats]bool{1: true, 5: true, 15: true},
	},
	beatDuration: 10,
	nextAfter:    55,
	expect:       150,
}, {
	track: drum.Track{
		Beats: [drum.NumBeats]bool{1: true, 5: true, 15: true},
	},
	beatDuration: 10,
	nextAfter:    150,
	expect:       150,
}, {
	track: drum.Track{
		Beats: [drum.NumBeats]bool{1: true, 5: true, 15: true},
	},
	beatDuration: 10,
	nextAfter:    151,
	expect:       170,
}, {
	track: drum.Track{
		Beats: [drum.NumBeats]bool{1: true, 5: true, 15: true},
	},
	beatDuration: 10,
	nextAfter:    165,
	expect:       170,
}, {
	track: drum.Track{
		Beats: [drum.NumBeats]bool{1: true, 5: true, 15: true},
	},
	beatDuration: 10,
	nextAfter:    170,
	expect:       170,
}, {
	track: drum.Track{
		Beats: [drum.NumBeats]bool{1: true, 5: true, 15: true},
	},
	beatDuration: 10,
	nextAfter:    171,
	expect:       210,
}}

func TestTrack(t *testing.T) {
	samples := []audio.Sample{1, 2, 3}
	for i, test := range trackTests {
		tr := &track{
			Track:        test.track,
			beatDuration: test.beatDuration,
			samples:      samples,
		}
		got, gotSamples := tr.nextAfter(test.nextAfter)
		if got != test.expect {
			t.Errorf("test %d: incorrect next time after %d, got %d, want %d", i, test.nextAfter, got, test.expect)
		}
		if &gotSamples[0] != &samples[0] {
			t.Errorf("wrong samples returned")
		}
	}
}

var sequencerTests = []struct {
	pattern            *drum.Pattern
	patches            map[string][]audio.Sample
	expectBeatDuration int64
	expect             []audio.Sample
}{{
	pattern: &drum.Pattern{
		Tempo: 120,
		Tracks: []drum.Track{{
			Name:  "a",
			Beats: [drum.NumBeats]bool{0: true},
		}, {
			Name:  "b",
			Beats: [drum.NumBeats]bool{1: true},
		}},
	},
	patches: map[string][]audio.Sample{
		"a": []audio.Sample{20, 18, 16, 14, 12, 10, 8},
		"b": []audio.Sample{21, 19, 15, 13, 11, 9, 7, 5},
	},
	expectBeatDuration: SampleRate / 2,
	expect: []audio.Sample{
		20, 18, 16, 14, 12,
		10 + 21, 8 + 19, 15, 13, 11,
		9, 7, 5, 0, 0,
		0, 0, 0, 0, 0,
		0, 0, 0, 0, 0,
		0, 0, 0, 0, 0,
		0, 0, 0, 0, 0,
		0, 0, 0, 0, 0,
		0, 0, 0, 0, 0,
		0, 0, 0, 0, 0,
		0, 0, 0, 0, 0,
		0, 0, 0, 0, 0,
		0, 0, 0, 0, 0,
		0, 0, 0, 0, 0,
		0, 0, 0, 0, 0,
		0, 0, 0, 0, 0,
		20, 18, 16, 14, 12,
		10 + 21, 8 + 19, 15, 13, 11,
		9, 7, 5, 0, 0,
	},
}}

func TestSequencer(t *testing.T) {
	for _, test := range sequencerTests {
		proc, err := New(test.pattern, test.patches)
		if err != nil {
			t.Fatalf("cannot make processor: %v", err)
		}
		seq := proc.(*sequencer)
		if got := getBeatDuration(seq); got != test.expectBeatDuration {
			t.Errorf("unexpected beat duration, got %d want %d", got, test.expectBeatDuration)
		}
		// Set the beat duration to 5 samples so we can test easily
		setBeatDuration(seq, 5)
		for i := 0; i < len(test.expect); i += 5 {
			out := make([]audio.Sample, 5)
			proc.Process(out)
			if !reflect.DeepEqual(out, test.expect[i:i+5]) {
				t.Errorf("frame %d, got %v want %v", i/5, out, test.expect[i:i+5])
			}
		}
	}
}

func getBeatDuration(seq *sequencer) int64 {
	d := int64(-1)
	for _, src := range seq.sources {
		sd := src.(*track).beatDuration
		if d == -1 {
			d = sd
		} else if sd != d {
			panic("inconsistent beat duration in tracks")
		}
	}
	return d
}

func setBeatDuration(seq *sequencer, d int64) {
	for _, src := range seq.sources {
		src.(*track).beatDuration = d
	}
}

// TODO test with silent tracks, silent patterns and drum sounds that aren't present.
