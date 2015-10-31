package snd

// Ring modulator
type Ring struct {
	*mono
	in0, in1 Sound
}

func NewRing(in0, in1 Sound) *Ring {
	return &Ring{newmono(nil), in0, in1}
}

func (ng *Ring) Inputs() []Sound {
	return []Sound{ng.in0, ng.in1}
}

func (ng *Ring) Prepare(uint64) {
	for i := range ng.out {
		if ng.off {
			ng.out[i] = 0
		} else {
			ng.out[i] = ng.in0.Sample(i) * ng.in1.Sample(i)
		}
	}
}
