package snd

import "testing"

func BenchmarkPan(b *testing.B) {
	pan := NewPan(0, newunit())
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		pan.Prepare(uint64(n))
	}
}
