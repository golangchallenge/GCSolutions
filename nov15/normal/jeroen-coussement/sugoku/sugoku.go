// Finds solution for a sudoku by constraint propagation.
package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
)

// grid represents the 9x9 sudoku field. Each position in the field is
// represented by a square
type grid [9][9]square

// square represents one position in the grid.
type square struct {
	peers []position
	units [][]position
	// map[int]bool is used to hold values. This way, we can easily test for p
	// presence by using values[val], which will return true if present and false
	// if absent
	values map[int]bool
}

// position represents the coordinates of a square in the grid
type position struct {
	row, col int
}

// getUnits returns a slice with 3 slices containing the positions of the row,
// column and box in which the position is present
func (p position) getUnits() [][]position {
	row := make([]position, 9)
	col := make([]position, 9)
	box := make([]position, 9)
	for i := 0; i < 9; i++ {
		row[i] = position{p.row, i}
		col[i] = position{i, p.col}
		box[i] = position{p.row/3*3 + i/3, p.col/3*3 + i%3}
	}
	return [][]position{row, col, box}
}

// getPeers returns a single slice in which each position of the position's row,
// column and box is present only once.
func (p position) getPeers() []position {
	// Use map instead of slice. This way we kan iterate the positions in the
	// units and add them to the map without worrying about adding the same
	// position twice. Maps don't keep order, but this is not necessary here.
	peers := make(map[position]struct{})
	for i := 0; i < 9; i++ {
		peers[position{p.row, i}] = struct{}{}
		peers[position{i, p.col}] = struct{}{}
		peers[position{p.row/3*3 + i/3, p.col/3*3 + i%3}] = struct{}{}
	}
	// delete current position
	delete(peers, p)

	var result []position
	for k := range peers {
		result = append(result, k)
	}

	return result
}

// makeDefaultGrid returns a grid which has all possible values on each position
func makeDefaultGrid() grid {
	var a = grid{}
	for r := range a {
		for c := range a[r] {
			a[r][c] = square{
				position{r, c}.getPeers(),
				position{r, c}.getUnits(),
				make(map[int]bool),
			}
			for i := 1; i <= 9; i++ {
				a[r][c].values[i] = true
			}
		}
	}
	return a
}

// assign assigns a values to a position in a grid by eliminating all other
// possible values on that position. Returns an error if a contradiction is
// detected.
func (g *grid) assign(p position, val int) error {
	// eliminate all other values on that position
	for v := range g[p.row][p.col].values {
		if v != val {
			if err := g.eliminate(p, v); err != nil {
				return err
			}
		}
	}
	return nil
}

// eliminate eliminates a value from the possible values in the specified
// position in the grid. After elimination, when there is only one possible
// value left on that position, in will eliminate this value from the peers.
// Also after elimination, we will check the positions units to see if there is
// only one more possibility for the value to be. If there is, we assign it
// there. Returns a error if a contradiction was found.
func (g *grid) eliminate(p position, val int) error {
	// get square on position p
	sq := g[p.row][p.col]
	if !sq.values[val] {
		// ok, already eliminated.
		return nil
	}
	// delete the value
	delete(sq.values, val)

	// if there is no value left, we have a contradiction
	if len(sq.values) == 0 {
		return errors.New("Contradiction was detected")
	}

	// If there is only one value left, eliminate it from peers.
	if len(sq.values) == 1 {
		for v := range sq.values {
			for _, peer := range sq.peers {
				if err := g.eliminate(peer, v); err != nil {
					return err
				}
			}
		}
	}

	// check if we can assign the value elsewhere in a unit, ie. there is only
	// one place the value can go.
	for _, u := range sq.units {
		var places []position
		for _, up := range u {
			if g[up.row][up.col].values[val] {
				places = append(places, up)
			}
		}
		// if no places found, we have a contradiction
		if len(places) == 0 {
			return errors.New("Contradiction was detected")
		}
		// if we found only one place, we can assign the value there.
		if len(places) == 1 {
			if err := g.assign(places[0], val); err != nil {
				return err
			}
		}
	}
	// all ok.
	return nil
}

// copy returns a new grid containing all the values of the grid it is called
// on.
func (g *grid) copy() grid {
	var a = grid{}
	for r := range g {
		for c, sq := range g[r] {
			a[r][c] = square{
				// peers and units can reference the same positions, as they are never
				// modified. No need to make new slices and copy the values.
				sq.peers,
				sq.units,
				// values do need to be copied.
				make(map[int]bool),
			}
			// copy values one by one.
			for k, v := range sq.values {
				a[r][c].values[k] = v
			}
		}
	}
	return a
}

// print pretty prints the grid it is called on.
func (g *grid) print() {
	for r := range g {
		for c := range g[r] {
			var vals []int
			for v := range g[r][c].values {
				vals = append(vals, v)
			}
			sort.Ints(vals)
			for _, v := range vals {
				fmt.Print(v)
			}
			fmt.Print(" ")
		}
		fmt.Print("\n")
	}
}

// search tries different possibilities until it finds a solution. It always starts
// on the possition having the least possibilities left.
func (g *grid) search(counter *int) error {
	// find positions of squares with most and least (but more then one if exists) values
	min := position{}
	max := position{}
	for r := range g {
		for c := range g[r] {
			if (len(g[r][c].values) < len(g[min.row][min.col].values) && len(g[r][c].values) > 1) || len(g[min.row][min.col].values) == 1 {
				min = position{r, c}
			}
			if len(g[r][c].values) > len(g[max.row][max.col].values) {
				max = position{r, c}
			}
		}
	}
	// If max value is 1, we have a solution
	if len(g[max.row][max.col].values) == 1 {
		return nil
	}

	// Increase counter, we are going to do an iteration.
	*counter++

	// Now, choose the square with the fewest possibilities and assign each value
	// once until we have a solution. The values assigned first is a randomly
	// chosen from the possible values, as maps don't keep order!
	for v := range g[min.row][min.col].values {
		g2 := g.copy()
		if err := g2.assign(min, v); err == nil {
			if err := g2.search(counter); err == nil {
				*g = g2
				return nil
			}
		}
	}
	return errors.New("No solution found")
}

// Solve take a sudoku challenge, assigns the values to the default grid and then
// searches for a solution.
func solve(sudoku []int) (g grid, iterations int, err error) {
	g = makeDefaultGrid()

	// assign the values of the sudoku challenge to the default grid...
	for i, v := range sudoku {
		if v != 0 {
			if err := g.assign(position{i / 9, i % 9}, v); err != nil {
				return grid{}, 0, err
			}
		}
	}

	// ... and find a solution!
	var counter int
	if err := g.search(&counter); err != nil {
		return grid{}, counter, err
	}
	return g, counter, nil

}

// rate gives a rating to the puzzle, based on the average iterations needed to
// solve it. Avg is calcuted from 20 solving cycles.
func rate(sudoku []int) (avgIterations float32, rating string, err error) {
	var total int
	for i := 0; i < 20; i++ {
		_, it, err := solve(sudoku)
		if err != nil {
			return 0, "", err
		}
		total += it
	}
	avgIterations = float32(total) / 20.0
	switch {
	case avgIterations < 10:
		rating = "easy"
		return
	case avgIterations < 20:
		rating = "medium"
		return
	}
	rating = "hard"
	return

}

// Parse input parses a slice of bytes representing the sudoku challenge and
// returns a slice of ints.
func parseInput(in []byte) ([]int, error) {
	// Use bytes.Map to eliminate all characters that are not ints or underscores.
	// Replace underscores by 0.
	res := bytes.Map(func(r rune) rune {
		if string(r) == "_" {
			return '0'
		} else if _, err := strconv.Atoi(string(r)); err != nil {
			return -1
		} else {
			return r
		}
	}, in)

	// The result should be 81 characters long. If not, the input was not valid.
	if len(res) != 81 {
		return nil, errors.New("Input was not in a valid format!")
	}

	// Convert the result in a slice of integers.
	ints := make([]int, 81)
	for i, r := range string(res) {
		// no need to check for errs, we made sure everything is an int during the
		// mapping earlier
		ints[i], _ = strconv.Atoi(string(r))
	}
	return ints, nil
}

func main() {
	// check if we need to rate the puzzle
	rateflag := flag.Bool("rate", false, "return a rating for this puzzle")
	flag.Parse()

	// Read from Stdin
	in, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	// parse the input the sudoku
	su, err := parseInput(in)
	if err != nil {
		fmt.Println("Invalid input!")
		fmt.Println(err)
		os.Exit(-1)
	}

	// solve
	solution, _, err := solve(su)

	if err != nil {
		fmt.Println("No solution found.")
		fmt.Println(err)
		os.Exit(-1)
	}
	fmt.Println("Solution found.")
	solution.print()

	if *rateflag {
		avg, rating, err := rate(su)
		if err != nil {
			fmt.Println("Error during rating :", err)
		} else {
			fmt.Println("Average iterations :", avg)
			fmt.Println("Rating :", rating)
		}
	}
}
