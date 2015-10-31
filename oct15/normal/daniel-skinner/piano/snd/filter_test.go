package snd

import "testing"

func BenchmarkLowPass(b *testing.B) {
	lp := NewLowPass(500, newunit())
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		lp.Prepare(uint64(n))
	}
}
