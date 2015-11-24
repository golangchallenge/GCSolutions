package main

import (
	"strings"
	"testing"

	"github.com/oijazsh/go-sudoku/sudoku"
)

func BenchmarkEasy(b *testing.B) {
	var grid sudoku.Grid
	s := `4 _ _ _ _ _ _ 5 _
  _ _ 3 _ _ 7 4 2 _
  _ 5 7 8 _ _ _ _ _
  3 4 _ 5 7 _ 2 _ 6
  _ _ 2 4 1 9 3 _ _
  8 _ 5 _ 6 2 _ 7 4
  _ _ _ _ _ 4 8 6 _
  _ 7 4 1 _ _ 9 _ _
  _ 2 _ _ _ _ _ _ 7`
	reader := strings.NewReader(s)
	grid.Write(reader)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g := grid
		g.Solve()
	}
}

func BenchmarkMedium(b *testing.B) {
	var grid sudoku.Grid
	s := `_ 9 2 _ 8 _ _ _ _
  3 _ 5 _ _ _ _ _ 8
  6 _ _ 5 _ _ 9 3 _
  _ _ _ _ 1 _ _ 2 _
  _ _ 9 3 2 7 6 _ _
  _ 4 _ _ 6 _ _ _ _
  _ 3 8 _ _ 4 _ _ 2
  9 _ _ _ _ _ 7 _ 5
  _ _ _ _ 7 _ 4 8 _`
	reader := strings.NewReader(s)
	grid.Write(reader)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g := grid
		g.Solve()
	}
}

func BenchmarkHard(b *testing.B) {
	var grid sudoku.Grid
	s := `2 _ _ 3 7 _ 6 _ _
  7 _ _ _ _ 5 _ 9 _
  4 _ _ _ _ 1 _ 3 _
  _ 9 _ _ _ _ _ 7 _
  3 _ _ _ 4 _ _ _ 1
  _ 6 _ _ _ _ _ 2 _
  _ 4 _ 6 _ _ _ _ 9
  _ 2 _ 1 _ _ _ _ 7
  _ _ 8 _ 2 7 _ _ 4`
	reader := strings.NewReader(s)
	grid.Write(reader)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g := grid
		g.Solve()
	}
}

func BenchmarkEvil(b *testing.B) {
	var grid sudoku.Grid
	s := `9 _ _ _ _ 8 4 7 _
  _ 1 _ _ _ _ _ _ 6
  _ _ 5 _ _ _ _ 2 _
  _ _ _ 9 _ _ 2 4 1
  _ _ _ 3 _ 4 _ _ _
  8 4 1 _ _ 7 _ _ _
  _ 5 _ _ _ _ 3 _ _
  1 _ _ _ _ _ _ 5 _
  _ 3 6 5 _ _ _ _ 9`
	reader := strings.NewReader(s)
	grid.Write(reader)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g := grid
		g.Solve()
	}
}

func BenchmarkSolveRank(b *testing.B) {
	var grid sudoku.Grid
	s := `9 _ _ _ _ 8 4 7 _
  _ 1 _ _ _ _ _ _ 6
  _ _ 5 _ _ _ _ 2 _
  _ _ _ 9 _ _ 2 4 1
  _ _ _ 3 _ 4 _ _ _
  8 4 1 _ _ 7 _ _ _
  _ 5 _ _ _ _ 3 _ _
  1 _ _ _ _ _ _ 5 _
  _ 3 6 5 _ _ _ _ 9`
	reader := strings.NewReader(s)
	grid.Write(reader)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g := grid
		g.SolveAndRank()
	}
}
