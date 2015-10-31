package snd

// TODO should mixer be stereo out?
// TODO perhaps this class is unnecessary, any sound could be a mixer
// if you can set multiple inputs, but might get confusing.
type Mixer struct {
	*mono
	ins []Sound
}

func NewMixer(ins ...Sound) *Mixer   { return &Mixer{newmono(nil), ins} }
func (mix *Mixer) Append(s ...Sound) { mix.ins = append(mix.ins, s...) }
func (mix *Mixer) Empty()            { mix.ins = nil }
func (mix *Mixer) Inputs() []Sound   { return mix.ins }

func (mix *Mixer) Prepare(uint64) {
	for i := range mix.out {
		mix.out[i] = 0
		if !mix.off {
			for _, in := range mix.ins {
				mix.out[i] += in.Sample(i)
			}
		}
	}
}
