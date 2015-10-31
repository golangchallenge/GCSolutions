package snd

import "time"

type Instrument struct {
	*mono
	tm int
}

func NewInstrument(in Sound) *Instrument {
	return &Instrument{newmono(in), 0}
}

func (nst *Instrument) OffIn(d time.Duration) {
	nst.tm = Dtof(d, nst.SampleRate())
}

func (nst *Instrument) On() {
	nst.tm = 0 // cancels any previous OffIn if not reached
	nst.mono.On()
}

func (nst *Instrument) Prepare(uint64) {
	for i := range nst.out {
		if nst.off {
			nst.out[i] = 0
		} else {
			nst.out[i] = nst.in.Sample(i)
		}

		if nst.tm > 0 {
			nst.tm--
			if nst.tm == 0 {
				nst.Off()
			}
		}
	}
}
