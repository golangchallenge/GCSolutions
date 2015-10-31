package snd

import (
	"log"
	"time"
)

// bufc implements a circular buffer.
type bufc struct {
	xs   []float64
	r, w int
}

// newbufc returns buffer of length n with read offset r.
func newbufc(n int, r int) *bufc { return &bufc{make([]float64, n), r, 0} }

func (b *bufc) read() (x float64) {
	x = b.xs[b.r]
	b.r++
	if b.r == len(b.xs) {
		b.r = 0
	}
	return
}

// readat returns x at r and next read position.
func (b *bufc) readat(r int) (x float64, n int) {
	n = r
	x = b.xs[n]
	n++
	if n == len(b.xs) {
		n = 0
	}
	return
}

func (b *bufc) write(x float64) (end bool) {
	b.xs[b.w] = x
	b.w++
	end = b.w == len(b.xs)
	if end {
		b.w = 0
	}
	return
}

// dtof converts time duration to approximate number of representative frames.
func Dtof(d time.Duration, sr float64) (f int) {
	return int(float64(d) / float64(time.Second) * sr)
}

// ftod converts f, number of frames, to approximate time duration.
func Ftod(f int, sr float64) (d time.Duration) {
	return time.Duration(float64(f) / sr * float64(time.Second))
}

type Delay struct {
	*mono
	line *bufc
}

// NewDelay returns Delay with sample buffer of a length approximated by d.
func NewDelay(d time.Duration, in Sound) *Delay {
	return &Delay{newmono(in), newbufc(Dtof(d, in.SampleRate()), 1)}
}

func (dly *Delay) Prepare(uint64) {
	for i := range dly.out {
		if dly.off {
			dly.out[i] = 0
		} else {
			dly.out[i] = dly.line.read()
			dly.line.write(dly.in.Sample(i))
		}
	}
}

// Tap is a tapped delay line, essentially a shorter delay within a larger one.
// TODO consider some type of method on Delay instead of a separate type.
// For example, Tap intentionally does not expose dly via Inputs() so why is it
// its own type? Conversely, that'd make Delay a mixer of sorts.
type Tap struct {
	*mono
	dly *Delay
	r   int
}

func NewTap(d time.Duration, in *Delay) *Tap {
	f := Dtof(d, in.SampleRate())
	n := len(in.line.xs)
	if f >= n {
		f = n - 1
	}
	// TODO this would fall out of sync if toggled off
	// separate from Delay suggesting a more intrinsic design
	// is required here.
	//
	// get ahead of the delay's write position
	r := in.line.w + (n - f)
	if r >= n {
		r -= n
	}
	return &Tap{newmono(nil), in, r}
}

func (tap *Tap) Prepare(uint64) {
	for i := range tap.out {
		if tap.off {
			tap.out[i] = 0
		} else {
			tap.out[i], tap.r = tap.dly.line.readat(tap.r)
		}
	}
}

type Comb struct {
	*mono
	line *bufc
	gain float64
}

func NewComb(gain float64, d time.Duration, in Sound) *Comb {
	return &Comb{newmono(in), newbufc(Dtof(d, in.SampleRate()), 1), gain}
}

func (cmb *Comb) Prepare(uint64) {
	for i := range cmb.out {
		if cmb.off {
			cmb.out[i] = 0
		} else {
			cmb.out[i] = cmb.line.read()
			cmb.line.write(cmb.in.Sample(i) + cmb.out[i]*cmb.gain)
		}
	}
}

// TODO cross-fade
type Loop struct {
	*mono
	line *bufc
	rec  bool

	syncrec bool
	sync    int
	syncpos int
}

func NewLoop(d time.Duration, in Sound) *Loop {
	return &Loop{newmono(in), newbufc(Dtof(d, in.SampleRate()), 0), false, false, 0, 0}
}

func NewLoopFrames(nframes int, in Sound) *Loop {
	return &Loop{newmono(in), newbufc(nframes, 0), false, false, 0, 0}
}

func (lp *Loop) SetBPM(bpm BPM) {
	lp.sync = Dtof(bpm.Dur(), lp.SampleRate())
}

func (lp *Loop) Syncing() bool { return lp.syncrec }

func (lp *Loop) Recording() bool { return lp.rec }

func (lp *Loop) Record() {
	// TODO may want to replace lp.line with new instance instead?
	// otherwise there will be left over data in buffer
	// if adding Stop() method, unless that's desireable?
	lp.line.w = 0
	if lp.sync == 0 {
		lp.rec = true
	} else {
		lp.syncpos = lp.sync
		lp.syncrec = true
	}
}

func (lp *Loop) Stop() {
	lp.rec = false
	lp.syncrec = false
	lp.syncpos = 0
	lp.line.r = 0
}

func (lp *Loop) Prepare(tc uint64) {
	nf := tc * uint64(lp.BufferLen())
	for i := range lp.out {
		lp.out[i] = 0
		if !lp.off {
			if lp.syncrec {
				if ((nf + uint64(i)) % uint64(lp.sync)) == 0 {
					lp.syncrec = false
					lp.rec = true
				} else {
					lp.syncpos--
					if lp.syncpos == 0 {
						log.Println("Failed to start loop.")
						lp.syncrec = false
					}
				}
			}

			if lp.rec {
				if done := lp.line.write(lp.in.Sample(i)); done {
					// lp.rec = false
					// lp.line.r = 0
					lp.Stop()
				}
			} else {
				lp.out[i] = lp.line.read()
			}
		}
	}
}
