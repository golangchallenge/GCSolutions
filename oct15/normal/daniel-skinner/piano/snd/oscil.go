package snd

type Oscil struct {
	*mono

	// TODO how much of this can I just make exported?
	// yes, freq and harm might not be thread safe exactly
	// but what's the worse that could happen if it was swapped out?
	h   Discrete
	idx float64

	freq    float64
	freqmod Sound

	amp    float64
	ampmod Sound

	phasemod Sound
}

func NewOscil(h Discrete, freq float64, freqmod Sound) *Oscil {
	return &Oscil{
		mono:    newmono(nil),
		h:       h,
		freq:    freq,
		freqmod: freqmod,
		amp:     DefaultAmpFac,
	}
}

func (osc *Oscil) SetFreq(hz float64, mod Sound) {
	osc.freq = hz
	osc.freqmod = mod
}

func (osc *Oscil) SetAmp(fac float64, mod Sound) {
	osc.amp = fac
	osc.ampmod = mod
}

func (osc *Oscil) SetPhase(mod Sound) {
	osc.phasemod = mod
}

func (osc *Oscil) Inputs() []Sound {
	return []Sound{osc.freqmod, osc.ampmod, osc.phasemod}
}

func (osc *Oscil) Prepare(tc uint64) {
	var (
		n float64 = float64(len(osc.h))
		f float64 = n / osc.sr
	)

	for i := range osc.out {
		if osc.off {
			osc.out[i] = 0
		} else {
			freq := osc.freq
			if osc.freqmod != nil {
				freq *= osc.freqmod.Sample(i)
			}

			amp := osc.amp
			if osc.ampmod != nil {
				amp *= osc.ampmod.Sample(i)
			}

			idx := 0
			if osc.phasemod != nil {
				idx = int(osc.idx+n*osc.phasemod.Sample(i)) & int(n-1)
			} else {
				idx = int(osc.idx) & int(n-1)
			}

			osc.out[i] = amp * osc.h[idx]
			osc.idx += freq * f
		}
	}
}
