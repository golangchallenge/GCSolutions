package snd

import (
	"testing"
	"time"
)

func BenchmarkADSR(b *testing.B) {
	ms := time.Millisecond
	env := NewADSR(5*ms, 10*ms, 15*ms, 20*ms, 0.7, 1, nil)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		env.Prepare(uint64(n))
	}
}

func BenchmarkDamp(b *testing.B) {
	env := NewDamp(50*time.Millisecond, newunit())
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		env.Prepare(uint64(n))
	}
}

func BenchmarkDrive(b *testing.B) {
	env := NewDrive(50*time.Millisecond, newunit())
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		env.Prepare(uint64(n))
	}
}
