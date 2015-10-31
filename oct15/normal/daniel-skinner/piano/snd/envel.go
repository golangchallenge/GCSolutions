package snd

import (
	"math"
	"time"
)

func ExpDecayFunc(t float64) float64 {
	return math.Exp(twopi * -t)
}

func ExpDecay() Discrete {
	sig := make(Discrete, 1024)
	sig.SampleFunc(ExpDecayFunc)
	return sig
}

func ExpDrive() Discrete {
	sig := ExpDecay()
	sig.Reverse()
	return sig
}

func LinearDecayFunc(t float64) float64 {
	return 1 - t
}

func LinearDecay() Discrete {
	sig := make(Discrete, 1024)
	sig.SampleFunc(LinearDecayFunc)
	return sig
}

func LinearDrive() Discrete {
	sig := LinearDecay()
	sig.Reverse()
	return sig
}

// TODO rename ...
type timed struct {
	sig Discrete
	nfr float64
}

func newtimed(sig Discrete, nfr int) *timed {
	return &timed{sig, float64(nfr)}
}

// TODO look into exposing this
type seq struct {
	*mono
	tms []*timed
	r   int
	pn  float64

	lk int
}

func newseq(in Sound) *seq {
	return &seq{mono: newmono(in), lk: -1}
}

func (sq *seq) Prepare(uint64) {
	for i := range sq.out {
		tm := sq.tms[sq.r]
		if sq.off {
			sq.out[i] = 0
		} else if sq.in == nil {
			sq.out[i] = tm.sig.SampleUnit(sq.pn / tm.nfr)
		} else {
			sq.out[i] = tm.sig.SampleUnit(sq.pn/tm.nfr) * sq.in.Sample(i)
		}

		// TODO finicky
		if sq.lk == sq.r {
			continue
		}

		sq.pn++
		if sq.pn >= tm.nfr {
			sq.pn = 0
			sq.r++
			if sq.r == len(sq.tms) {
				sq.r = 0
			}
		}
	}
}

// TODO reimplement sustain functionality
type ADSR struct {
	*seq
	sustaining bool
}

func NewADSR(attack, decay, sustain, release time.Duration, susamp, maxamp float64, in Sound) *ADSR {
	adsr := &ADSR{newseq(in), false}
	sr := adsr.SampleRate()

	atksig := LinearDrive()
	atksig.NormalizeRange(0, maxamp)
	atk := newtimed(atksig, Dtof(attack, sr))

	dcysig := LinearDecay()
	dcysig.NormalizeRange(maxamp, susamp)
	dcy := newtimed(dcysig, Dtof(decay, sr))

	sus := newtimed(Discrete{susamp, susamp}, Dtof(sustain, sr))

	relsig := ExpDecay()
	relsig.NormalizeRange(susamp, 0)
	rel := newtimed(relsig, Dtof(release, sr))

	adsr.tms = []*timed{atk, dcy, sus, rel}
	return adsr
}

func (adsr *ADSR) Dur() time.Duration {
	var n int
	for _, tm := range adsr.tms {
		n += int(tm.nfr)
	}
	return Ftod(n, adsr.SampleRate())
}

// Restart resets envelope to start from attack period.
func (adsr *ADSR) Restart() { adsr.r, adsr.pn, adsr.lk = 0, 0, -1 }

// Sustain locks envelope when sustain period is reached.
func (adsr *ADSR) Sustain() {
	adsr.sustaining = true
	adsr.lk = 2
}

// Release immediately releases envelope from anywhere and starts release period.
func (adsr *ADSR) Release() (ok bool) {
	adsr.sustaining = false
	adsr.lk = -1
	if ok = adsr.r < 3; ok {
		adsr.r, adsr.pn = 3, 0
	}
	return
}

type Damp struct {
	*mono
	sig  Discrete
	i, n float64
}

func NewDamp(d time.Duration, in Sound) *Damp {
	sd := newmono(in)
	return &Damp{
		mono: sd,
		sig:  ExpDecay(),
		n:    float64(Dtof(d, sd.SampleRate())),
	}
}

func (dmp *Damp) Prepare(uint64) {
	for i := range dmp.out {
		if dmp.off {
			dmp.out[i] = 0
		} else if dmp.in == nil {
			dmp.out[i] = dmp.sig.SampleUnit(dmp.i / dmp.n)
		} else {
			dmp.out[i] = dmp.in.Sample(i) * dmp.sig.SampleUnit(dmp.i/dmp.n)
		}
		dmp.i++
		if dmp.i == dmp.n {
			dmp.i = 0
		}
	}
}

type Drive struct {
	*mono
	sig  Discrete
	i, n float64
}

func NewDrive(d time.Duration, in Sound) *Drive {
	sd := newmono(in)
	return &Drive{
		mono: sd,
		sig:  ExpDrive(),
		n:    float64(Dtof(d, sd.SampleRate())),
	}
}

func (drv *Drive) Prepare(uint64) {
	for i := range drv.out {
		if drv.off {
			drv.out[i] = 0
		} else if drv.in == nil {
			drv.out[i] = drv.sig.SampleUnit(drv.i / drv.n)
		} else {
			drv.out[i] = drv.in.Sample(i) * drv.sig.SampleUnit(drv.i/drv.n)
		}
		drv.i++
		if drv.i == drv.n {
			drv.i = 0
		}
	}
}
