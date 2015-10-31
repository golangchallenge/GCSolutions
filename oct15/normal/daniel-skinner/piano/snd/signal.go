package snd

import (
	"fmt"
	"math"
)

const twopi = 2 * math.Pi

// ContinuousFunc represents a continuous signal over infinite time.
type ContinuousFunc func(t float64) float64

// Discrete represents a discrete signal over its length as time.
type Discrete []float64

// SampleFunc samples a ContinuousFunc over time belonging to [0..1].
// If sig is nil, space will be allocated.
// Length of sig must be a power of 2 or Sample will panic.
func (sig *Discrete) SampleFunc(fn ContinuousFunc) {
	if *sig == nil {
		*sig = make(Discrete, DefaultBufferLen)
	}
	if n := len(*sig); n == 0 || n&(n-1) != 0 {
		panic(fmt.Errorf("Discrete len(%v) not a power of 2", n))
	}
	n := float64(len(*sig))
	for i := 0.0; i < n; i++ {
		(*sig)[int(i)] = fn(i / n)
	}
}

// SampleUnit accepts value t belonging to [0..1] and returns corresponding value
// from sig by length.
func (sig *Discrete) SampleUnit(t float64) float64 {
	if t > 1 {
		t = 1
	}
	if t < 0 {
		t = 0
	}
	n := float64(len(*sig) - 1)
	i := int(t * n)
	return (*sig)[i]
}

func (sig *Discrete) Sample(i int) float64 {
	return (*sig)[i&(len(*sig)-1)]
}

// Normalize alters sig so values belong to [-1..1].
func (sig *Discrete) Normalize() {
	var max float64
	for _, x := range *sig {
		a := math.Abs(x)
		if max < a {
			max = a
		}
	}
	for i, x := range *sig {
		(*sig)[i] = x / max
	}
	(*sig)[len(*sig)-1] = (*sig)[0]
	// sig.NormalizeRange(-1, 1)
}

// NormalizeRange alters sig so values belong to [s..e].
func (sig *Discrete) NormalizeRange(s, e float64) {
	if s > e {
		s, e = e, s
	}
	n := e - s

	var min, max float64
	for _, x := range *sig {
		if min > x {
			min = x
		}
		if max < x {
			max = x
		}
	}

	for i, x := range *sig {
		pct := (x - min) / (max - min)
		(*sig)[i] = s + pct*n
	}
	(*sig)[len(*sig)-1] = (*sig)[0]
}

func (sig *Discrete) Reverse() {
	for l, r := 0, len(*sig)-1; l < r; l, r = l+1, r-1 {
		(*sig)[l], (*sig)[r] = (*sig)[r], (*sig)[l]
	}
}

// Add performs additive synthesis from the fundamental, a, for the partial harmonic, pth.
func (sig *Discrete) Add(a Discrete, pth int) {
	for i := range *sig {
		j := i * pth % (len(a) - 1)
		(*sig)[i] += a[j] * (1 / float64(pth))
	}
}

// SineFunc is the continuous signal of a sine wave.
func SineFunc(t float64) float64 {
	return math.Sin(twopi * t)
}

// Sine returns a discrete sample of SineFunc.
func Sine() (sig Discrete) {
	sig.SampleFunc(SineFunc)
	return
}

// TriangleFunc is the continuous signal of a triangle wave.
func TriangleFunc(t float64) float64 {
	return 2*math.Abs(SawtoothFunc(t)) - 1
}

// Triangle returns a discrete sample of TriangleFunc.
func Triangle() (sig Discrete) {
	sig.SampleFunc(TriangleFunc)
	return
}

// SquareFunc is the continuous signal of a square wave.
func SquareFunc(t float64) float64 {
	if math.Signbit(math.Sin(twopi * t)) {
		return -1
	}
	return 1
}

// Square returns a discrete sample of SquareFunc.
func Square() (sig Discrete) {
	sig.SampleFunc(SquareFunc)
	return
}

// SawtoothFunc is the continuous signal of a sawtooth wave.
func SawtoothFunc(t float64) float64 {
	return 2 * (t - math.Floor(0.5+t))
}

// Sawtooth returns a discrete sample of SawtoothFunc.
func Sawtooth() (sig Discrete) {
	sig.SampleFunc(SawtoothFunc)
	return
}

// fundamental default used for sinusoidal synthesis.
var fundamental = Sine()

// SquareSynthesis adds odd partial harmonics belonging to [3..n], creating a sinusoidal wave.
func SquareSynthesis(n int) Discrete {
	sig := Sine()
	for i := 3; i <= n; i += 2 {
		sig.Add(fundamental, i)
	}
	sig.Normalize()
	return sig
}

// SawtoothSynthesis adds all partial harmonics belonging to [2..n], creating a sinusoidal wave
// that is the inverse of a sawtooth.
func SawtoothSynthesis(n int) Discrete {
	sig := Sine()
	for i := 2; i <= n; i++ {
		sig.Add(fundamental, i)
	}
	sig.Normalize()
	return sig
}
