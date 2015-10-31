package snd

import "testing"

func BenchmarkDispatchSerial(b *testing.B) {
	sd := mksound()
	inps := GetInputs(sd)

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		tc := uint64(n)
		for _, inp := range inps {
			inp.sd.Prepare(tc)
		}
	}
}

func TestDispatch(t *testing.T) {
	sd := mksound()
	inps := GetInputs(sd)

	tc := uint64(1)
	dp := new(Dispatcher)

	// TODO better test than "does it hang?"
	dp.Dispatch(tc, inps...)
}

func BenchmarkDispatch(b *testing.B) {
	sd := mksound()
	inps := GetInputs(sd)

	dp := new(Dispatcher)

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		dp.Dispatch(uint64(n), inps...)
	}
}

func TestGetInputs(t *testing.T) {
	sd := mksound()
	inps := GetInputs(sd)
	if len(inps) <= 1 {
		t.Fatal("getinps did not produce a result")
	}
	last := inps[0].wt
	for i, inp := range inps {
		if inp.wt > last {
			t.Fatal("inputs are not sorted highest to lowest")
		}
		last = inp.wt
		t.Log(i, inp)
	}
}

func BenchmarkGetInputs(b *testing.B) {
	sd := mksound()
	var inps []*Input
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		inps = GetInputs(sd)
	}
	_ = inps
}

func TestByWTSlice(t *testing.T) {
	sd := mksound()
	inps := GetInputs(sd)
	sl := ByWT(inps).Slice()

	want := len(inps)
	total := 0
	for i, p := range sl {
		total += len(p)
		t.Log(i, len(p))
	}
	if total != want {
		t.Fatalf("Have length %v, want %v", total, want)
	}
}
