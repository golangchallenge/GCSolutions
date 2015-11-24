package main

import (
	"bufio"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGridGetUnits(t *testing.T) {
	p := position{0, 0}
	units := p.getUnits()

	assert.True(t, len(units) == 3, "Units must contain 3 elements")

	for _, u := range units {
		assert.True(t, len(u) == 9, "Each Unit must contain 9 elements")
	}
}

func TestGridGetPeers(t *testing.T) {
	p := position{0, 0}
	peers := p.getPeers()

	assert.True(t, len(peers) == 20, "Peers must contain 20 elements")
}

func TestMakeDefaultGrid(t *testing.T) {
	g := makeDefaultGrid()

	for r := range g {
		for c := range g[r] {
			for i := 1; i <= 9; i++ {
				assert.True(t, g[r][c].values[i], "All positions must contain all values")
			}
		}
	}
}

func TestGridCopy(t *testing.T) {
	g1 := makeDefaultGrid()
	g2 := g1.copy()

	for r := range g1 {
		for c := range g1[r] {
			assert.True(t, &g1[r][c].peers[0] == &g2[r][c].peers[0], "the slices for the peers and units in g2 must point to the same values as g1")
			for v := range g1[r][c].values {
				assert.True(t, g2[r][c].values[v], "g2 should contain all values of g1")
			}
		}
	}
}

func TestGridAssign(t *testing.T) {
	g := makeDefaultGrid()
	// we will assign a value of 1 to postition {0,0} of the default grid. This
	// means that afterwards, {0,0} must only containt 1 element, 1, and all peers
	// must have value 1 removed.

	g.assign(position{0, 0}, 1)
	assert.True(t, len(g[0][0].values) == 1 && g[0][0].values[1], "{0,0} must only containt 1 element and it must be 1.")

	for _, p := range g[0][0].peers {
		assert.True(t, len(g[p.row][p.col].values) == 8 && !g[p.row][p.col].values[1], "All peers must now containt 8 elements and should not contain 1")
	}

	err := g.assign(position{0, 1}, 1)
	assert.NotNil(t, err, "1 cannot be assigned to {0,1} after it has been assigned to {0,0} ")
}

func TestGridEliminate(t *testing.T) {
	g := makeDefaultGrid()
	// we will eliminate a values of 1 at postition {0,0} of the default grid. This
	// means that afterwards, {0,0} must only containt 8 elements, and 1 should
	// not be part of it.

	g.eliminate(position{0, 0}, 1)
	assert.True(t, len(g[0][0].values) == 8 && !g[0][0].values[1], "{0,0} must containt 8 element and and should not containt 1.")

	for r := range g {
		for c := range g[r] {
			if r != 0 && c != 0 {
				assert.True(t, len(g[r][c].values) == 9, "all other positions other than {0,0} should still have all values")
			}
		}
	}

	for i := 2; i <= 9; i++ {
		err := g.eliminate(position{0, 0}, i)
		if i == 9 {
			assert.NotNil(t, err, "Cannot eliminate all values from a position")
		}
	}
}

func TestGridSearch(t *testing.T) {
	g := makeDefaultGrid()
	su := [...]int{
		1, 0, 3, 0, 0, 6, 0, 8, 0,
		0, 5, 0, 0, 8, 0, 1, 2, 0,
		7, 0, 9, 1, 0, 3, 0, 5, 6,
		0, 3, 0, 0, 6, 7, 0, 9, 0,
		5, 0, 7, 8, 0, 0, 0, 3, 0,
		8, 0, 1, 0, 3, 0, 5, 0, 7,
		0, 4, 0, 0, 7, 8, 0, 1, 0,
		6, 0, 8, 0, 0, 2, 0, 4, 0,
		0, 1, 2, 0, 4, 5, 0, 7, 8,
	}

	for i, v := range su {
		if v != 0 {
			err := g.assign(position{i / 9, i % 9}, v)
			assert.True(t, err == nil, "Should be able to assign all values of this challenge.")
		}
	}

	var counter int
	err := g.search(&counter)
	assert.True(t, err == nil, "Searching a solution should not return an err for this challenge.")

	for i, v := range su {
		assert.True(t, len(g[i/9][i%9].values) == 1, "Sudoku should be solved! Each position should only have one value.")
		if v != 0 {
			assert.True(t, g[i/9][i%9].values[v], "Values of the challenge should be present on their position in the solution")
		}
	}
}

func TestSolve(t *testing.T) {
	su := [...]int{
		1, 0, 3, 0, 0, 6, 0, 8, 0,
		0, 5, 0, 0, 8, 0, 1, 2, 0,
		7, 0, 9, 1, 0, 3, 0, 5, 6,
		0, 3, 0, 0, 6, 7, 0, 9, 0,
		5, 0, 7, 8, 0, 0, 0, 3, 0,
		8, 0, 1, 0, 3, 0, 5, 0, 7,
		0, 4, 0, 0, 7, 8, 0, 1, 0,
		6, 0, 8, 0, 0, 2, 0, 4, 0,
		0, 1, 2, 0, 4, 5, 0, 7, 8,
	}

	_, _, err := solve(su[:])
	assert.True(t, err == nil, "Challenge should be solved.")

}

func TestRate(t *testing.T) {
	su := [...]int{
		1, 0, 3, 0, 0, 6, 0, 8, 0,
		0, 5, 0, 0, 8, 0, 1, 2, 0,
		7, 0, 9, 1, 0, 3, 0, 5, 6,
		0, 3, 0, 0, 6, 7, 0, 9, 0,
		5, 0, 7, 8, 0, 0, 0, 3, 0,
		8, 0, 1, 0, 3, 0, 5, 0, 7,
		0, 4, 0, 0, 7, 8, 0, 1, 0,
		6, 0, 8, 0, 0, 2, 0, 4, 0,
		0, 1, 2, 0, 4, 5, 0, 7, 8,
	}

	_, _, err := rate(su[:])
	assert.True(t, err == nil, "Rating should not give errors.")

}

func BenchmarkSolve50HardSudokus(b *testing.B) {
	file, err := os.Open("./challenges/lists/hard.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var challenges [][]int
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		chal, err := parseInput(scanner.Bytes())
		if err == nil {
			challenges = append(challenges, chal)
		}
	}

	for i := 0; i < b.N; i++ {
		for _, c := range challenges {
			_, _, err := solve(c)
			if err != nil {
				panic(err)
			}
		}
	}
}

func BenchmarkSolve50EasySudokus(b *testing.B) {
	file, err := os.Open("./challenges/lists/easy.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var challenges [][]int
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		chal, err := parseInput(scanner.Bytes())
		if err == nil {
			challenges = append(challenges, chal)
		}
	}

	for i := 0; i < b.N; i++ {
		for _, c := range challenges {
			_, _, err := solve(c)
			if err != nil {
				panic(err)
			}
		}
	}
}
