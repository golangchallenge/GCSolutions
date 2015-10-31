package snd

import "testing"

func BenchmarkMixer(b *testing.B) {
	mix := NewMixer()
	for i := 0; i < 2; i++ {
		mix.Append(newunit())
	}
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		mix.Prepare(uint64(n))
	}
}
