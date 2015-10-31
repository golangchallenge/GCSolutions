package snd

import "testing"

var resf float64

func BenchmarkDiscrete(b *testing.B) {
	sig := Sine()
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for i := 0; i < 256; i++ {
			resf = sig.Sample(n)
		}
	}
}

type foo [256]float64

func BenchmarkDiscreteCopy(b *testing.B) {
	sig := Sine()
	dup := make(Discrete, len(sig))
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		copy(dup, sig)
	}
}
