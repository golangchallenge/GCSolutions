package main

import (
	"testing"
)

var bigBoxPile = make([]uint32, 1000)

// Pretty much worst case. All branches get hit.
func BenchmarkPack_1x1s(b *testing.B) {
	r := repacker{}
	for n := 0; n < b.N; n++ {
		r.boxes1x1 = bigBoxPile
		r.PackPallet()
	}
}

// Very fast. Best case.
func BenchmarkPack_4x4s(b *testing.B) {
	r := repacker{}
	for n := 0; n < b.N; n++ {
		r.boxes4x4 = bigBoxPile
		r.PackPallet()
	}
}

func BenchmarkPack_4x1s(b *testing.B) {
	r := repacker{}
	for n := 0; n < b.N; n++ {
		r.boxes4x1 = bigBoxPile
		r.PackPallet()
	}
}

func BenchmarkPack_4x2s(b *testing.B) {
	r := repacker{}
	for n := 0; n < b.N; n++ {
		r.boxes4x2 = bigBoxPile
		r.PackPallet()
	}
}

func BenchmarkPack_4x3s(b *testing.B) {
	r := repacker{}
	for n := 0; n < b.N; n++ {
		r.boxes4x3 = bigBoxPile
		r.boxes2x1 = bigBoxPile
		r.PackPallet()
	}
}

func BenchmarkPack_3x1s(b *testing.B) {
	r := repacker{}
	for n := 0; n < b.N; n++ {
		r.boxes3x1 = bigBoxPile
		r.PackPallet()
	}
}

func BenchmarkPack_Mixed(b *testing.B) {
	r := repacker{}
	//1111
	//1111
	//2233
	//2245
	b42 := []uint32{1}
	b22 := []uint32{2}
	b21 := []uint32{3}
	b11 := []uint32{4, 5}
	for n := 0; n < b.N; n++ {
		r.boxes4x2 = b42
		r.boxes2x2 = b22
		r.boxes2x1 = b21
		r.boxes1x1 = b11
		r.PackPallet()
	}
}

//make a collection of 100 or so trucks to use for tests. Hopefully will be somewhat representative.
var trucksToUnload = []*truck{}

func init() {
	for i := 0; i < 100; i++ {
		trucksToUnload = append(trucksToUnload, gentruck())
	}
}

func BenchmarkUnpack(b *testing.B) {
	r := repacker{}
	for n := 0; n < b.N; n++ {
		r.Unload(trucksToUnload[n%len(trucksToUnload)])
	}
}
