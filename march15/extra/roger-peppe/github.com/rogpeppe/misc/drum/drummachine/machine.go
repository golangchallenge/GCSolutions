// Package drummachine acts as a Sigourney audio source that
// can play drum samples.
//
// See http://github.com/nf/sigourney for more information
// on the Sigourney audio synthesizer.
package drummachine

import (
	"fmt"

	"github.com/nf/sigourney/audio"

	"github.com/rogpeppe/misc/drum"
)

// SampleRate holds the sample rate used by the audio samples,
// in samples per second.
const SampleRate = 44100

// New returns a new drum machine module that will repeatedly play
// the drum pattern p using the given patch samples.
func New(p *drum.Pattern, patches map[string][]audio.Sample) (audio.Processor, error) {
	var seq sequencer
	beatDuration := int64(SampleRate/(p.Tempo/60) + 0.5)
	for _, tr := range p.Tracks {
		if !hasBeat(tr) {
			// Ignore silent track.
			continue
		}
		patch := patches[tr.Name]
		if len(patch) == 0 {
			return nil, fmt.Errorf("drum sound %q not found", tr.Name)
		}
		seq.sources = append(seq.sources, &track{
			Track:        tr,
			beatDuration: beatDuration,
			samples:      patch,
		})
	}
	if len(seq.sources) == 0 {
		// Whereof one cannot speak, thereof one must be silent.
		return silence{}, nil
	}
	return &seq, nil
}

type silence struct{}

func (silence) Process(out []audio.Sample) {
	zero(out)
}

// hasBeat reports whether the given track
// has any drum beats.
func hasBeat(tr drum.Track) bool {
	for _, on := range tr.Beats {
		if on {
			return true
		}
	}
	return false
}

// sequencer sequences a set of sources,
// mixing together their results by addition.
type sequencer struct {
	// sources holds all the sources for the sequencer.
	sources []source

	// current holds all the samples that are currently
	// playing.
	current [][]audio.Sample

	// t holds the current sample time.
	t int64

	// next holds the next time that any of the sources
	// are scheduled to play something new.
	next int64
}

// source represents a source of samples.
type source interface {
	// nextAfter returns the next time that a sample should be
	// played after (or equal to) the given time t.
	nextAfter(t int64) (int64, []audio.Sample)
}

// processn processes n samples into out.
// It updates seq.t and seq.current.
func (seq *sequencer) processn(out []audio.Sample, n int) {
	zero(out[0:n])
	remove := false
	for i, samples := range seq.current {
		n := n
		if n >= len(samples) {
			remove = true
			n = len(samples)
		}
		for i, sample := range samples[0:n] {
			out[i] += sample
		}
		seq.current[i] = samples[n:]
	}
	seq.t += int64(n)

	if !remove {
		return
	}
	j := 0
	for _, samples := range seq.current {
		if len(samples) != 0 {
			seq.current[j] = samples
			j++
		}
	}
	seq.current = seq.current[0:j]
}

const maxInt64 = int64(0x7fffffffffffffff)

func (seq *sequencer) Process(out []audio.Sample) {
	for len(out) > 0 {
		if seq.t == seq.next {
			next := maxInt64
			for _, source := range seq.sources {
				// If we have lots of sources, we could use a heap.
				sourceNext, sourceSamples := source.nextAfter(seq.t)
				if sourceNext == seq.t {
					seq.current = append(seq.current, sourceSamples)
					sourceNext, sourceSamples = source.nextAfter(seq.t + 1)
				}
				if sourceNext < next {
					next = sourceNext
				}
			}
			if next == maxInt64 {
				panic("no source returned a next event")
			}
			seq.next = next
		}
		n := seq.next - seq.t
		if n > int64(len(out)) {
			n = int64(len(out))
		}
		seq.processn(out, int(n))
		out = out[n:]
	}
}

func zero(s []audio.Sample) {
	for i := range s {
		s[i] = 0
	}
}

// track implements the source interface for a drum pattern track.
type track struct {
	drum.Track
	beatDuration int64
	samples      []audio.Sample
}

func (tr *track) nextAfter(t int64) (int64, []audio.Sample) {
	// Calculate the current offset into the beats
	// of the current time, rounding up so we won't
	// find a beat in the past.
	beatStart := (t + tr.beatDuration - 1) / tr.beatDuration
	// Note: we assume there is at least one drum sound in the track,
	// something that's checked when we create the machine.
	for i := beatStart; ; i++ {
		if tr.Beats[i%int64(len(tr.Beats))] {
			return i * tr.beatDuration, tr.samples
		}
	}
}
