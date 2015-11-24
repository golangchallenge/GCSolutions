// sudoji implements a lightweight sudoku solver built around a non-deterministic in-place Las Vegas algorithm
// that can solve any valid puzzle not only without backtracking but without any forwardtracking as well. However,
// this functionality comes at the expense of an inability to recognize all but the most obvious of invalid inputs.
package main

import (
	"fmt"
	"io/ioutil"
	"math/big"
	"math/rand"
	"os"
	"time"
)

// A Grid represents a Sudoku puzzle in grid form, using the digit 0 for unfilled cells.
type Grid [9][9]int

var (
	puzzle   Grid
	unknowns []int
)

func main() {
	input, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}
	d := make([]int, 0, 81)
	for _, b := range input {
		// Ignore input bytes except for digits, '_', and '.'
		switch {
		case b == '_' || b == '.' || b == '0':
			d = append(d, 0)
		case '1' <= b && b <= '9':
			d = append(d, int(b-'0'))
		}
	}
	if len(d) != 81 {
		panic(fmt.Sprintf("Invalid input: length %v", len(d)))
	}

	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			puzzle[r][c] = d[r*9+c]
		}
	}
	if Conflict(puzzle) {
		panic("Puzzle cannot be solved")
	}
	unknowns = Unknowns(puzzle)

	rand.Seed(time.Now().UnixNano())

	for n, one := big.NewInt(0), big.NewInt(1); ; n.Add(n, one) {
		if g := ProposeSolution(puzzle); !Conflict(g) {
			Output(g)
			fmt.Println("Difficulty: ", Difficulty(n))
			break
		}
	}
}

// Unknowns returns a slice of all cell values missing from the input Grid.
func Unknowns(g Grid) []int {
	count := make([]int, 10)
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			count[puzzle[r][c]]++
		}
	}
	u := make([]int, 0, count[0])
	for digit := 1; digit <= 9; digit++ {
		for c := count[digit]; c < 9; c++ {
			u = append(u, digit)
		}
	}
	return u
}

// ProposeSolution utilizes a non-deterministic in-place Las Vegas algorithm to generate and return a completed Grid that may be a valid solution to the input Grid.
func ProposeSolution(g Grid) Grid {
	i, p := 0, rand.Perm(len(unknowns))
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if g[r][c] == 0 {
				g[r][c] = unknowns[p[i]]
				i++
			}
		}
	}
	return g
}

// Conflict reports whether a same-digit conflict exists in the input Grid.
func Conflict(g Grid) bool {
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if g[r][c] == 0 {
				continue
			}
			for i := 1; i < 9; i++ {
				switch g[r][c] {
				case g[(r+i)%9][c]:
					return true
				case g[r][(c+i)%9]:
					return true
				case g[r/3*3+(r%3*3+c%3+i)/3%3][c/3*3+(r%3*3+c%3+i)%3]:
					return true
				}
			}
		}
	}
	return false
}

// Output writes a Grid to standard output.
func Output(g Grid) {
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			fmt.Printf("%v ", g[r][c])
		}
		fmt.Printf("\n")
	}
}

// Difficulty returns a string description of the difficulty of a Grid solution based on the logarithm of the number of steps required to produce it.
func Difficulty(n *big.Int) string {
	d := n.BitLen() / 3
	if d < 6 {
		return []string{"piece of cake", "easy", "pretty easy", "medium", "pretty hard", "hard"}[d]
	}
	d -= 6
	a0 := []string{"quite ", "very ", "very very ", "really ", "really very ", "really really "}
	a1 := []string{"", "extremely ", "incredibly ", "exceedingly ", "exquisitely ", "unbelievably ", "unimaginably ", "stupendously ", "tremendously ", "inconceivably ", "mind-bogglingly ", "incomprehensibly ", "earth-shatteringly "}
	var r string
	for e := len(a0) * len(a1); e <= d; d -= e {
		r += "ridiculously "
	}
	return a0[d%len(a0)] + r + a1[(d/len(a0))%len(a1)] + "hard"
}
