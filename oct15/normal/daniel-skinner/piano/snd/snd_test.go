package snd

import (
	"testing"
	"time"
)

type unit struct{ *mono }

func newunit() *unit {
	u := &unit{newmono(nil)}
	for i := range u.out {
		u.out[i] = DefaultAmpFac
	}
	return u
}

func (u *unit) Prepare(uint64) {}

type zeros struct{ *mono }

func newzeros() *zeros { return &zeros{newmono(nil)} }

func (z *zeros) Prepare(uint64) {
	for i := range z.out {
		if z.off {
			z.out[i] = 0
		} else {
			z.out[i] = 1
		}
	}
}

func BenchmarkZeros(b *testing.B) {
	z := newzeros()
	b.ReportAllocs()
	b.ResetTimer()
	for n := 1; n <= b.N; n++ {
		z.Prepare(uint64(n))
	}
}

// mksound returns a 12-key piano synth as might be found in an app.
func mksound() Sound {
	mix := NewMixer()
	for i := 0; i < 12; i++ {
		oscil := NewOscil(Sawtooth(), 440, NewOscil(Sine(), 2, nil))
		oscil.SetPhase(NewOscil(Square(), 200, nil))

		comb := NewComb(0.8, 10*time.Millisecond, oscil)
		adsr := NewADSR(50*time.Millisecond, 500*time.Millisecond, 100*time.Millisecond, 350*time.Millisecond, 0.4, 1, comb)
		instr := NewInstrument(adsr)
		mix.Append(instr)
	}
	loop := NewLoop(5*time.Second, mix)
	mixloop := NewMixer(mix, loop)
	lp := NewLowPass(1500, mixloop)
	mixwf, err := NewWaveform(nil, 4, lp)
	if err != nil {
		panic(err)
	}
	return NewPan(0, mixwf)
}

// func mkthoughtsaboutcreatingstuff() Sound {
// What if these were Options passed in ?
// How many things actually have options?

// Would need a way that an "option" and an "instance" could
// interchangably be used here, such as implementing the same interface?
// or having some type of lazy initialization that would allow easy access
// to the instance pointers.

// sine := Sine()
// sawtooth := Sawtooth(4)
// square := Square(4)

// phaser := snd.Oscil{
// Harm: square,
// Freq: 200,
// }.New()

// osc := snd.Oscil{
// Harm:  sawtooth,
// Freq:  440,
// Mod:   snd.Oscil{Harm: sine, Freq: 2}, // every new osc has independent mod
// Phase: phaser,                         // every new osc reuses same phaser
// }.New()

// cmb := snd.Comb{
// Gain: 0.8,
// Dur:  10 * time.Millisecond,
// }.New()

// env := snd.ADSR{
// Attack:  50 * time.Millisecond,
// Decay:   500 * time.Millisecond,
// Sustain: 100 * time.Millisecond,
// Release: 350 * time.Millisecond,
// SusAmp:  0.4,
// MaxAmp:  1,
// }.New()

// cmb.SetInput(osc)
// env.SetInput(cmb)

// this is just bad for so many reasons.
// &Pan{
// Left:  1,
// Right: 1,
// In: &Waveform{
// Buf: 4,
// In: &LowPass{
// Cutoff: 1500,
// In: &Mixer{
// Ins: []Sound{
// &Loop{
// Duration: 5,
// },
// &Mixer{
// Ins: []Sound{}
// },
// },
// },
// },
// },
// }
// }

func TestDecibel(t *testing.T) {
	tests := []struct {
		db  Decibel
		amp float64
	}{
		{0, 1},
		{1, 1.1220},
		{3, 1.4125},
		{6, 1.9952},
		{10, 3.1622},
	}

	for _, test := range tests {
		if !equals(test.db.Amp(), test.amp) {
			t.Errorf("%s have %v, want %v", test.db, test.db.Amp(), test.amp)
		}
	}
}
