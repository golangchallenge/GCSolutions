package snd

import "math"

// LowPass is a 3rd order IIR filter.
//
// Recursive implementation of the Gaussian filter.
type LowPass struct {
	*mono

	// normalization factor
	b float64
	// coefficients
	b0, b1, b2, b3 float64
	// delays
	d1, d2, d3 float64

	// TODO eek, temporary
	passthrough bool
}

func (lp *LowPass) SetPassthrough(b bool) { lp.passthrough = b }
func (lp *LowPass) Passthrough() bool     { return lp.passthrough }

func NewLowPass(freq float64, in Sound) *LowPass {
	q := 5.0
	s := in.SampleRate() / freq / q

	if s > 2.5 {
		q = 0.98711*s - 0.96330
	} else {
		q = 3.97156 - 4.14554*math.Sqrt(1-0.26891*s)
	}

	q2 := q * q
	q3 := q * q * q

	// redefined from paper to (1 / b0) to save an op div during prepare.
	b0 := 1 / (1.57825 + 2.44413*q + 1.4281*q2 + 0.422205*q3)
	b1 := 2.44413*q + 2.85619*q2 + 1.26661*q3
	b2 := -(1.4281*q2 + 1.26661*q3)
	b3 := 0.422205 * q3
	b := 1 - ((b1 + b2 + b3) * b0)

	b1 *= b0
	b2 *= b0
	b3 *= b0

	return &LowPass{mono: newmono(in), b: b, b0: b0, b1: b1, b2: b2, b3: b3}
}

func (lp *LowPass) Prepare(uint64) {
	for i, x := range lp.in.Samples() {
		if lp.off {
			lp.out[i] = 0
		} else if lp.passthrough {
			lp.out[i] = x
		} else {
			lp.out[i] = lp.b*x + lp.b1*lp.d1 + lp.b2*lp.d2 + lp.b3*lp.d3
		}
		lp.d3, lp.d2, lp.d1 = lp.d2, lp.d1, lp.out[i]
	}
}
