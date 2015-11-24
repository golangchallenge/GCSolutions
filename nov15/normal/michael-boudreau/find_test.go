package main

import (
	"testing"
)

func BenchmarkSingleRankedCoordinate_LR_00(b *testing.B) {
	finder := &RankedCoordFinder{}
	for n := 0; n < b.N; n++ {
		finder.NextOpenCoordinate(testBoardLR, &Coord{0, 0})
	}
}
func BenchmarkSingleRankedCoordinate_RL_00(b *testing.B) {
	finder := &RankedCoordFinder{}
	for n := 0; n < b.N; n++ {
		finder.NextOpenCoordinate(testBoardRL, &Coord{0, 0})
	}
}
func BenchmarkSingleRankedCoordinate_Full_00(b *testing.B) {
	finder := &RankedCoordFinder{}
	for n := 0; n < b.N; n++ {
		finder.NextOpenCoordinate(testBoardFull, &Coord{0, 0})
	}
}
func BenchmarkSingleRankedCoordinate_Empty_00(b *testing.B) {
	finder := &RankedCoordFinder{}
	for n := 0; n < b.N; n++ {
		finder.NextOpenCoordinate(testBoardEmpty, &Coord{0, 0})
	}
}
func BenchmarkSingleRankedCoordinate_LR_88(b *testing.B) {
	finder := &RankedCoordFinder{}
	for n := 0; n < b.N; n++ {
		finder.NextOpenCoordinate(testBoardLR, &Coord{8, 8})
	}
}
func BenchmarkSingleRankedCoordinate_RL_88(b *testing.B) {
	finder := &RankedCoordFinder{}
	for n := 0; n < b.N; n++ {
		finder.NextOpenCoordinate(testBoardRL, &Coord{8, 8})
	}
}
func BenchmarkSingleRankedCoordinate_Full_88(b *testing.B) {
	finder := &RankedCoordFinder{}
	for n := 0; n < b.N; n++ {
		finder.NextOpenCoordinate(testBoardFull, &Coord{8, 8})
	}
}
func BenchmarkSingleRankedCoordinate_Empty_88(b *testing.B) {
	finder := &RankedCoordFinder{}
	for n := 0; n < b.N; n++ {
		finder.NextOpenCoordinate(testBoardEmpty, &Coord{8, 8})
	}
}

func BenchmarkSingleClosestCoordinate_LR_00(b *testing.B) {
	finder := &ClosestCoordFinder{}
	for n := 0; n < b.N; n++ {
		finder.NextOpenCoordinate(testBoardLR, &Coord{0, 0})
	}
}
func BenchmarkSingleClosestCoordinate_RL_00(b *testing.B) {
	finder := &ClosestCoordFinder{}
	for n := 0; n < b.N; n++ {
		finder.NextOpenCoordinate(testBoardRL, &Coord{0, 0})
	}
}
func BenchmarkSingleClosestCoordinate_Full_00(b *testing.B) {
	finder := &ClosestCoordFinder{}
	for n := 0; n < b.N; n++ {
		finder.NextOpenCoordinate(testBoardFull, &Coord{0, 0})
	}
}
func BenchmarkSingleClosestCoordinate_Empty_00(b *testing.B) {
	finder := &ClosestCoordFinder{}
	for n := 0; n < b.N; n++ {
		finder.NextOpenCoordinate(testBoardEmpty, &Coord{0, 0})
	}
}
func BenchmarkSingleClosestCoordinate_LR_88(b *testing.B) {
	finder := &ClosestCoordFinder{}
	for n := 0; n < b.N; n++ {
		finder.NextOpenCoordinate(testBoardLR, &Coord{8, 8})
	}
}
func BenchmarkSingleClosestCoordinate_RL_88(b *testing.B) {
	finder := &ClosestCoordFinder{}
	for n := 0; n < b.N; n++ {
		finder.NextOpenCoordinate(testBoardRL, &Coord{8, 8})
	}
}
func BenchmarkSingleClosestCoordinate_Full_88(b *testing.B) {
	finder := &ClosestCoordFinder{}
	for n := 0; n < b.N; n++ {
		finder.NextOpenCoordinate(testBoardFull, &Coord{8, 8})
	}
}
func BenchmarkSingleClosestCoordinate_Empty_88(b *testing.B) {
	finder := &ClosestCoordFinder{}
	for n := 0; n < b.N; n++ {
		finder.NextOpenCoordinate(testBoardEmpty, &Coord{8, 8})
	}
}

var testBoardEmpty = Board([][]byte{
	{0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0},
})
var testBoardFull = Board([][]byte{
	{1, 2, 3, 4, 5, 6, 7, 8, 9},
	{4, 5, 6, 7, 8, 9, 1, 2, 3},
	{7, 8, 9, 1, 2, 3, 4, 5, 6},
	{2, 3, 4, 5, 6, 7, 8, 9, 1},
	{5, 6, 7, 8, 9, 1, 2, 3, 4},
	{8, 9, 1, 2, 3, 4, 5, 6, 7},
	{9, 4, 5, 6, 7, 8, 3, 1, 2},
	{6, 7, 8, 3, 1, 2, 9, 4, 5},
	{3, 1, 2, 9, 4, 5, 6, 7, 8},
})
var testBoardLR = Board([][]byte{
	{1, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 2, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 3, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 4, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 5, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 6, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 7, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 8, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 9},
})
var testBoardRL = Board([][]byte{
	{0, 0, 0, 0, 0, 0, 0, 0, 9},
	{0, 0, 0, 0, 0, 0, 0, 8, 0},
	{0, 0, 0, 0, 0, 0, 7, 0, 0},
	{0, 0, 0, 0, 0, 6, 0, 0, 0},
	{0, 0, 0, 0, 5, 0, 0, 0, 0},
	{0, 0, 0, 4, 0, 0, 0, 0, 0},
	{0, 0, 3, 0, 0, 0, 0, 0, 0},
	{0, 2, 0, 0, 0, 0, 0, 0, 0},
	{1, 0, 0, 0, 0, 0, 0, 0, 0},
})
var testFinderBoardOne = Board([][]byte{
	{1, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 2, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 3, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 4, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 5, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 6, 0, 0, 0},
	{2, 5, 1, 0, 0, 0, 7, 4, 3},
	{4, 3, 7, 0, 0, 0, 6, 8, 2},
	{6, 8, 9, 0, 0, 0, 5, 1, 0},
})
