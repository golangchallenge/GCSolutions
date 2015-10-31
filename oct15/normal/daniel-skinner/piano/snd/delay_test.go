package snd

import (
	"testing"
	"time"
)

func TestBufc(t *testing.T) {
	for _, td := range []struct{ n, r int }{
		{10, 1},
		{4, 2},
		{7, 3},
		{5, 4},
	} {
		n, r := td.n, td.r
		buf := newbufc(n, r)
		for i := 1; i <= 100; i++ {
			if buf.write(float64(i)) && i%n != 0 {
				t.Fatalf("buf write incorrectly signaled end [i=%v]", i)
			}
			if 0 > buf.w || buf.w >= n {
				t.Fatalf("buf write position out of bounds %+v", buf)
			}

			if x := buf.read(); i <= (n-r) && x != 0 {
				t.Fatalf("buf read %v during first pass, want 0 [i=%v] %+v", x, i, buf)
			} else if w := i - n + r; i > (n-r) && int(x) != w {
				t.Fatalf("buf read %v, want %v [i=%v] %+v", x, w, i, buf)
			}
			if 0 > buf.r || buf.r >= n {
				t.Fatalf("buf read position out of bounds %+v", buf)
			}
		}
	}
}

func TestDtof(t *testing.T) {
	sr := 44100.0
	eps := time.Duration(1 / sr * float64(time.Second)) // 1Hz as time.Duration
	d := 75 * time.Millisecond
	x := Ftod(Dtof(d, sr), sr)
	if diff := d - x; diff > eps {
		t.Fatalf("%s greater than epsilon %s", diff, eps)
	}
}

func BenchmarkDelay(b *testing.B) {
	dly := NewDelay(100*time.Millisecond, newunit())
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		dly.Prepare(uint64(n))
	}
}

func BenchmarkDelayTap(b *testing.B) {
	dly := NewDelay(100*time.Millisecond, newunit())
	tap := NewTap(50*time.Millisecond, dly)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		dly.Prepare(uint64(n))
		tap.Prepare(uint64(n))
	}
}

func BenchmarkComb(b *testing.B) {
	cmb := NewComb(0.8, 100*time.Millisecond, newunit())
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		cmb.Prepare(uint64(n))
	}
}
