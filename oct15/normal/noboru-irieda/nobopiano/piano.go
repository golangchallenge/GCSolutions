package main

type Piano struct {
	notes     []bool
	oscilator Oscilator
}

func New(freqs []float32) *Piano {
	p := new(Piano)
	p.notes = make([]bool, len(freqs))
	envelopes := []Oscilator{}
	for i, f := range freqs {
		osc := Multiplex(
			G(0.4, GenOscilator(f)),   // base sin wave
			G(0.1, GenOscilator(f+2)), // vibrate effect
			G(0.2, GenOscilator(f*2)), // 2 times freq
			G(0.1, GenOscilator(f*4)), // 3 times freq
		)
		envelopes = append(envelopes, G(0.5, GenEnvelope(&p.notes[i], osc)))
	}
	p.oscilator = Multiplex(envelopes...) // all note oscilator multiplex
	return p
}

func (p *Piano) NoteOn(key int) {
	p.notes[key] = true
}

func (p *Piano) NoteOff(key int) {
	p.notes[key] = false
}
func (p *Piano) GetOscilator() Oscilator { return p.oscilator }
