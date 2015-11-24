package main

import (
	"bytes"
	"strings"
	"testing"
)

// Source: https://github.com/attractivechaos/plb/blob/365a033abafecc90931a9a56e63e5c32c0ca68b4/sudoku/sudoku.txt
//
// Note: The repository at the above link seems to contain no LICENSE or similar file.
const hard20 = `
..............3.85..1.2.......5.7.....4...1...9.......5......73..2.1........4...9	near worst case for brute-force solver (wiki)
.......12........3..23..4....18....5.6..7.8.......9.....85.....9...4.5..47...6...	gsf's sudoku q1 (Platinum Blonde)
.2..5.7..4..1....68....3...2....8..3.4..2.5.....6...1...2.9.....9......57.4...9..	(Cheese)
........3..1..56...9..4..7......9.5.7.......8.5.4.2....8..2..9...35..1..6........	(Fata Morgana)
12.3....435....1....4........54..2..6...7.........8.9...31..5.......9.7.....6...8	(Red Dwarf)
1.......2.9.4...5...6...7...5.9.3.......7.......85..4.7.....6...3...9.8...2.....1	(Easter Monster)
.......39.....1..5..3.5.8....8.9...6.7...2...1..4.......9.8..5..2....6..4..7.....	Nicolas Juillerat's Sudoku explainer 1.2.1 (top 5)
12.3.....4.....3....3.5......42..5......8...9.6...5.7...15..2......9..6......7..8
..3..6.8....1..2......7...4..9..8.6..3..4...1.7.2.....3....5.....5...6..98.....5.
1.......9..67...2..8....4......75.3...5..2....6.3......9....8..6...4...1..25...6.
..9...4...7.3...2.8...6...71..8....6....1..7.....56...3....5..1.4.....9...2...7..
....9..5..1.....3...23..7....45...7.8.....2.......64...9..1.....8..6......54....7	dukuso's suexrat9 (top 1)
4...3.......6..8..........1....5..9..8....6...7.2........1.27..5.3....4.9........	from http://magictour.free.fr/topn87 (top 3)
7.8...3.....2.1...5.........4.....263...8.......1...9..9.6....4....7.5...........
3.7.4...........918........4.....7.....16.......25..........38..9....5...2.6.....
........8..3...4...9..2..6.....79.......612...6.5.2.7...8...5...1.....2.4.5.....3	dukuso's suexratt (top 1)
.......1.4.........2...........5.4.7..8...3....1.9....3..4..2...5.1........8.6...	first 2 from sudoku17
.......12....35......6...7.7.....3.....4..8..1...........12.....8.....4..5....6..
1.......2.9.4...5...6...7...5.3.4.......6........58.4...2...6...3...9.8.7.......1	2 from http://www.setbb.com/phpbb/viewtopic.php?p=10478
.....1.2.3...4.5.....6....7..2.....1.8..9..3.4.....8..5....2....9..3.4....67.....
`

var hardGames []game

func init() {
	for _, v := range strings.Split(hard20, "\n") {
		if v = strings.TrimSpace(v); v == "" {
			continue
		}

		var g game
		for i, n := range v[:81] {
			if n != '.' {
				g.set(i/9, i%9, int(n-'0'))
			}
		}
		hardGames = append(hardGames, g)
	}
}

var easyBoard = [9][9]int{
	{1, 0, 3, 0, 0, 6, 0, 8, 0},
	{0, 5, 0, 0, 8, 0, 1, 2, 0},
	{7, 0, 9, 1, 0, 3, 0, 5, 6},
	{0, 3, 0, 0, 6, 7, 0, 9, 0},
	{5, 0, 7, 8, 0, 0, 0, 3, 0},
	{8, 0, 1, 0, 3, 0, 5, 0, 7},
	{0, 4, 0, 0, 7, 8, 0, 1, 0},
	{6, 0, 8, 0, 0, 2, 0, 4, 0},
	{0, 1, 2, 0, 4, 5, 0, 7, 8},
}

var easySolution = [9][9]int{
	{1, 2, 3, 4, 5, 6, 7, 8, 9},
	{4, 5, 6, 7, 8, 9, 1, 2, 3},
	{7, 8, 9, 1, 2, 3, 4, 5, 6},
	{2, 3, 4, 5, 6, 7, 8, 9, 1},
	{5, 6, 7, 8, 9, 1, 2, 3, 4},
	{8, 9, 1, 2, 3, 4, 5, 6, 7},
	{3, 4, 5, 6, 7, 8, 9, 1, 2},
	{6, 7, 8, 9, 1, 2, 3, 4, 5},
	{9, 1, 2, 3, 4, 5, 6, 7, 8},
}

func (g *game) verify() bool {
	for _, row := range g.rows {
		if row != full {
			return false
		}
	}
	for _, col := range g.cols {
		if col != full {
			return false
		}
	}
	for _, row := range g.boxes {
		for _, box := range row {
			if box != full {
				return false
			}
		}
	}
	return true
}

func TestEasy(t *testing.T) {
	var g game
	for y, row := range easyBoard {
		for x, v := range row {
			if v != 0 {
				g.set(y, x, v)
			}
		}
	}

	var buf bytes.Buffer
	g.writeTo(&buf)
	t.Logf("problem, %v clues\n%s", g.finished, buf.Bytes())

	if g, e := g.solve() && g.verify(), true; g != e {
		t.Fatal(g, e)
	}

	buf.Reset()
	g.writeTo(&buf)
	t.Logf("solution\n%s", buf.Bytes())

	if g, e := g.board, easySolution; g != e {
		t.Fatal("wrong solution")
	}
}

func TestHard20(t *testing.T) {
	for i, g := range hardGames {
		var buf bytes.Buffer
		g.writeTo(&buf)
		t.Logf("problem %v, %v clues\n%s", i+1, g.finished, buf.Bytes())

		if g, e := g.solve() && g.verify(), true; g != e {
			t.Fatal(i, g, e)
		}

		buf.Reset()
		g.writeTo(&buf)
		t.Logf("solution\n%s", buf.Bytes())
	}
}

func BenchmarkEasy(b *testing.B) {
	var g, g0 game
	for y, row := range easyBoard {
		for x, v := range row {
			if v != 0 {
				g0.set(y, x, v)
			}
		}
	}
	for i := 0; i < b.N; i++ {
		g = g0
		if !g.solve() {
			b.Fatal("no solution found")
		}
	}
}

func BenchmarkHard1To20(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, g := range hardGames {
			if !g.solve() {
				b.Fatal("no solution found")
			}
		}
	}
}

func benchmarkHard(b *testing.B, n int) {
	for i := 0; i < b.N; i++ {
		g := hardGames[n]
		if !g.solve() {
			b.Fatal("no solution found")
		}
	}
}

func BenchmarkHard1(b *testing.B)  { benchmarkHard(b, 0) }
func BenchmarkHard2(b *testing.B)  { benchmarkHard(b, 1) }
func BenchmarkHard3(b *testing.B)  { benchmarkHard(b, 2) }
func BenchmarkHard4(b *testing.B)  { benchmarkHard(b, 3) }
func BenchmarkHard5(b *testing.B)  { benchmarkHard(b, 4) }
func BenchmarkHard6(b *testing.B)  { benchmarkHard(b, 5) }
func BenchmarkHard7(b *testing.B)  { benchmarkHard(b, 6) }
func BenchmarkHard8(b *testing.B)  { benchmarkHard(b, 7) }
func BenchmarkHard9(b *testing.B)  { benchmarkHard(b, 8) }
func BenchmarkHard10(b *testing.B) { benchmarkHard(b, 9) }
func BenchmarkHard11(b *testing.B) { benchmarkHard(b, 10) }
func BenchmarkHard12(b *testing.B) { benchmarkHard(b, 11) }
func BenchmarkHard13(b *testing.B) { benchmarkHard(b, 12) }
func BenchmarkHard14(b *testing.B) { benchmarkHard(b, 13) }
func BenchmarkHard15(b *testing.B) { benchmarkHard(b, 14) }
func BenchmarkHard16(b *testing.B) { benchmarkHard(b, 15) }
func BenchmarkHard17(b *testing.B) { benchmarkHard(b, 16) }
func BenchmarkHard18(b *testing.B) { benchmarkHard(b, 17) }
func BenchmarkHard19(b *testing.B) { benchmarkHard(b, 18) }
func BenchmarkHard20(b *testing.B) { benchmarkHard(b, 19) }
