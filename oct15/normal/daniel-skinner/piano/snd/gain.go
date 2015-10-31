package snd

type Gain struct {
	*mono
	a float64
}

func NewGain(a float64, in Sound) *Gain {
	return &Gain{newmono(in), a}
}

func (gn *Gain) SetAmp(a float64) {
	gn.a = a
}

func (gn *Gain) Prepare(uint64) {
	for i, x := range gn.in.Samples() {
		if gn.off {
			gn.out[i] = 0
		} else {
			gn.out[i] = gn.a * x
		}
	}
}
