package snd

import (
	"testing"
	"time"
)

func TestRingcopy(t *testing.T) {
	xs := []float64{1, 2, 3, 4}
	tds := []struct {
		dn   int
		off  int
		ret  int
		want []float64
	}{
		{2, 2, 0, []float64{3, 4}},
		{8, 2, 2, []float64{3, 4, 1, 2, 3, 4, 1, 2}},
	}

	for i, td := range tds {
		have := make([]float64, td.dn)
		r := ringcopy(have, xs, td.off)
		if r != td.ret {
			t.Errorf("tds[%v] returned %v, want %v", i, r, td.ret)
		}
		for j, x := range td.want {
			if x != have[j] {
				t.Errorf("tds[%v] have %+v, want %+v", i, have, td.want)
				break
			}
		}
	}
}

func BenchmarkRingcopy(b *testing.B) {
	// src := []float64(ExpDecay())
	src := make(Discrete, 32)
	src.SampleFunc(ExpDecayFunc)
	dst := make([]float64, 256)
	r := 0
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		r = ringcopy(dst, src, r)
	}
}

func BenchmarkFreeze(b *testing.B) {
	frz := NewFreeze(1*time.Second, newunit())
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		frz.Prepare(uint64(n + 1))
	}
}
