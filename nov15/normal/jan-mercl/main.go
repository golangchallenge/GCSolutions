package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

var popcnt [1024]int

func init() {
	for i := range popcnt {
		bits := 0
		for j := i; j != 0; j >>= 1 {
			if j&1 != 0 {
				bits++
			}
		}
		popcnt[i] = bits
	}
}

const full = 0x3fe // 11 1111 1110, set of 1-9.

type set int

func (s *set) add(n int)     { *s = *s | 1<<uint(n) }
func (s set) has(n int) bool { return s&(1<<uint(n)) != 0 }
func (s *set) remove(n int)  { *s &^= 1 << uint(n) }

type game struct {
	finished int
	board    [9][9]int
	rows     [9]set
	cols     [9]set
	boxes    [3][3]set
}

func (g *game) writeTo(w io.Writer) {
	for _, row := range g.board {
		for _, cell := range row {
			switch cell {
			case 0:
				fmt.Fprint(w, "_ ")
			default:
				fmt.Fprintf(w, "%v ", cell)
			}
		}
		fmt.Fprintln(w)
	}
}

func (g *game) reset(y, x int) {
	v := g.board[y][x]
	g.rows[y].remove(v)
	g.cols[x].remove(v)
	g.boxes[y/3][x/3].remove(v)
	g.finished--
	g.board[y][x] = 0
}

func (g *game) set(y, x, v int) {
	if old := g.board[y][x]; old != 0 {
		g.rows[y].remove(old)
		g.cols[x].remove(old)
		g.boxes[y/3][x/3].remove(old)
		g.finished--
	}

	g.rows[y].add(v)
	g.cols[x].add(v)
	g.boxes[y/3][x/3].add(v)
	g.finished++
	g.board[y][x] = v
}

func (g *game) solve() bool {
	if g.finished == 81 {
		return true
	}

	var best set
	var bestN, bestX, bestY int
outer:
	for y, row := range g.rows {
		if row == full {
			continue
		}

		for x, col := range g.cols {
			if col == full || g.board[y][x] != 0 {
				continue
			}

			bits := row | col | g.boxes[y/3][x/3]
			if n := popcnt[bits]; n > bestN {
				if bits == full {
					return false
				}

				bestN, best, bestX, bestY = n, bits, x, y
				if n == 8 {
					break outer
				}
			}
		}
	}
	for i := 1; i < 10; i++ {
		if !best.has(i) {
			g.set(bestY, bestX, i)
			if g.solve() {
				return true
			}
		}
	}

	g.reset(bestY, bestX)
	return false
}

func main() {
	if err := main1(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func main1() error {
	var g game
	scanner := bufio.NewScanner(os.Stdin)
	for y := 0; scanner.Scan(); y++ {
		if y == 9 {
			return fmt.Errorf("too many input lines")
		}

		s0 := scanner.Text()
		s := strings.Replace(s0, " ", "", -1)
		if len(s) != 9 {
			return fmt.Errorf("invalid input line: %q\n", s0)
		}

		for x, r := range s {
			switch {
			case r == '_':
				// OK, nop.
			case r >= '1' && r <= '9':
				g.set(y, x, int(r-'0'))
			default:
				return fmt.Errorf("invalid input character: %c\n", r)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading standard input: %v", err)
	}

	if !g.solve() {
		return fmt.Errorf("no solution exists")
	}

	g.writeTo(os.Stdout)
	return nil
}
