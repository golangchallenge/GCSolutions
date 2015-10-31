package snd

import "math"

var (
	onesqrt2 = 1 / math.Sqrt(2)

	panres float64 = 512
	panfac [1024]float64
)

func init() {
	for i := range panfac {
		n := float64(i)/panres - 1
		panfac[i] = onesqrt2 * (1 - n) / math.Sqrt(1+(n*n))
	}
}

func getpanfac(xf float64) float64 {
	if xf > 1 {
		xf = 1
	} else if xf < -1 {
		xf = -1
	}
	return panfac[int(panres*(1+xf))]
}

type Pan struct {
	*stereo
	xf float64
}

func NewPan(xf float64, in Sound) *Pan {
	return &Pan{newstereo(in), xf}
}

// SetAmount sets amount an input is panned across two outputs where amt belongs to [-1..1].
func (pan *Pan) SetAmount(xf float64) { pan.xf = xf }

// Prepare interleaves the left and right channels.
func (pan *Pan) Prepare(uint64) {
	for i, x := range pan.in.Samples() {
		if pan.l.off {
			pan.l.out[i] = 0
		} else {
			pan.l.out[i] = x * getpanfac(pan.xf)
		}
		if pan.r.off {
			pan.r.out[i] = 0
		} else {
			pan.r.out[i] = x * getpanfac(-pan.xf)
		}
		pan.out[i*2] = pan.l.out[i]
		pan.out[i*2+1] = pan.r.out[i]
	}
}
