package snd

import "testing"

func BenchmarkOscil(b *testing.B) {
	osc := NewOscil(Sine(), 440, nil)
	// dp := new(Dispatcher)
	// inps := GetInputs(osc)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		osc.Prepare(uint64(n))
		// dp.Dispatch(uint64(n), inps...)
	}
}

func BenchmarkOscilMod(b *testing.B) {
	osc := NewOscil(Sine(), 440, NewOscil(Sine(), 2, nil))
	inps := GetInputs(osc)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, inp := range inps {
			inp.sd.Prepare(uint64(n))
		}
	}
}

func BenchmarkOscilAmp(b *testing.B) {
	osc := NewOscil(Sine(), 440, nil)
	osc.SetAmp(1, NewOscil(Sine(), 2, nil))
	inps := GetInputs(osc)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, inp := range inps {
			inp.sd.Prepare(uint64(n))
		}
	}
}

func BenchmarkOscilPhase(b *testing.B) {
	osc := NewOscil(Sine(), 440, nil)
	osc.SetPhase(NewOscil(Sine(), 2, nil))
	inps := GetInputs(osc)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, inp := range inps {
			inp.sd.Prepare(uint64(n))
		}
	}
}

func BenchmarkOscilAll(b *testing.B) {
	osc := NewOscil(Sine(), 440, NewOscil(Sine(), 2, nil))
	osc.SetAmp(1, NewOscil(Sine(), 2, nil))
	osc.SetPhase(NewOscil(Sine(), 2, nil))
	// dp := new(Dispatcher)
	inps := GetInputs(osc)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		// dp.Dispatch(uint64(n), inps...)
		for _, inp := range inps {
			inp.sd.Prepare(uint64(n))
		}
	}
}

func BenchmarkOscilReuse(b *testing.B) {
	mod := NewOscil(Sine(), 2, nil)
	osc := NewOscil(Sine(), 440, mod)
	osc.SetAmp(1, mod)
	osc.SetPhase(mod)
	inps := GetInputs(osc)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, inp := range inps {
			inp.sd.Prepare(uint64(n))
		}
	}
}
