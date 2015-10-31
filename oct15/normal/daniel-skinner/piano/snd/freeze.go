package snd

import (
	"math"
	"time"
)

type Freeze struct {
	*mono
	sig, prv Discrete
	r        int
}

func NewFreeze(d time.Duration, in Sound) *Freeze {
	f := Dtof(d, in.SampleRate())

	n := f
	if n == 0 || n&(n-1) != 0 {
		_, e := math.Frexp(float64(n))
		n = int(math.Ldexp(1, e))
	}

	frz := &Freeze{mono: newmono(nil), prv: make(Discrete, n)}
	frz.sig = frz.prv[:f]

	inps := GetInputs(in)
	dp := new(Dispatcher)

	// t := time.Now()
	for i := 0; i < n; i += in.BufferLen() {
		dp.Dispatch(1, inps...)
		ringcopy(frz.sig[i:i+in.BufferLen()], in.Samples(), 0)
	}
	// log.Println("freeze took", time.Now().Sub(t))
	return frz
}

func (frz *Freeze) Restart() { frz.r = 0 }

var empty = make([]float64, 256)

func ringcopy(dst, src []float64, r int) int {
	dn, sn := len(dst), len(src)
	for w := 0; w < dn; {
		x := copy(dst[w:], src[r:])
		w += x
		r += x
		if r == sn {
			r = 0
		}
	}
	return r
}

func (frz *Freeze) Off() {
	frz.mono.Off()
	copy(frz.out, empty)
}

func (frz *Freeze) Prepare(uint64) {
	if frz.off {
		n := len(frz.out)
		frz.r = (frz.r + n) & (n - 1)
	} else {
		frz.r = ringcopy(frz.out, frz.sig, frz.r)
	}

	// for i := range frz.out {
	// if frz.off {
	// frz.out[i] = 0
	// } else {
	// frz.out[i] = frz.sig[frz.pos]
	// frz.pos++
	// if frz.pos == frz.nfr {
	// frz.pos = 0
	// }
	// }
	// }
}
