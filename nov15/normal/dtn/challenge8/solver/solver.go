// solver provides a simple sudoku solver
package solver

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
)

const (
	BOX = 3
	DIM = 9
	ALL = (1 << DIM) - 1
)

var (
	ErrInvalidInput     = errors.New("solver: invalid input")
	ErrMalformedInput   = errors.New("solver: malformed input")
	ErrNoSolution       = errors.New("solver: no solution")
	ErrMultipleSolution = errors.New("solver: more than one solution exist")
)

func singleton(v uint8) uint16 {
	return 1 << (v - 1)
}

// bc counts the number of 1's in the binary representation of v
func bc(v uint16) uint8 {
	var cnt uint8
	for {
		if v == 0 {
			return cnt
		}
		v &= (v - 1)
		cnt++
	}
}

type Board [9][9]uint8

func (src *Board) copy() *Board {
	dst := new(Board)
	for r := 0; r < DIM; r++ {
		for c := 0; c < DIM; c++ {
			dst[r][c] = src[r][c]
		}
	}
	return dst
}

func (b *Board) String() string {
	var repr string
	for r := 0; r < DIM; r++ {
		repr += fmt.Sprintln(b[r])
	}
	return repr
}

type Candidates [9][9]uint16

func (src *Candidates) copy() *Candidates {
	dst := new(Candidates)
	for r := 0; r < DIM; r++ {
		for c := 0; c < DIM; c++ {
			dst[r][c] = src[r][c]
		}
	}
	return dst
}

func (cand *Candidates) eliminate(r, c int, v uint8) {
	mask := ^singleton(v)
	saved := cand[r][c]
	defer func() {
		cand[r][c] = saved
	}()

	for ri := 0; ri < DIM; ri++ {
		cand[ri][c] &= mask
	}

	for ci := 0; ci < DIM; ci++ {
		cand[r][ci] &= mask
	}

	brow, bcol := r/BOX, c/BOX
	for i := 0; i < BOX; i++ {
		for j := 0; j < BOX; j++ {
			cand[brow*BOX+i][bcol*BOX+j] &= mask
		}
	}
}

func (input *Board) setCandidates() *Candidates {
	cand := new(Candidates)
	for ri := 0; ri < DIM; ri++ {
		for ci := 0; ci < DIM; ci++ {
			cand[ri][ci] = ALL
		}
	}

	for ri := 0; ri < DIM; ri++ {
		for ci := 0; ci < DIM; ci++ {
			if input[ri][ci] == 0 {
				continue
			}
			cand[ri][ci] = singleton(input[ri][ci])
			cand.eliminate(ri, ci, input[ri][ci])
		}
	}

	return cand
}

func isValid(problem *Board, cand *Candidates) bool {
	for ri := 0; ri < DIM; ri++ {
		for ci := 0; ci < DIM; ci++ {
			if problem[ri][ci] == 0 {
				continue
			}
			if (singleton(problem[ri][ci]) & cand[ri][ci]) == 0 {
				return false
			}
		}
	}
	return true
}

type Cell struct {
	r, c int
}

type Status int

const (
	solved Status = iota
	unsolvable
	notSolvedYet
)

func fewestCandidates(problem *Board, candidates *Candidates) (Cell, Status) {
	r, c := -1, -1
	var count uint8 = DIM + 1

	for ri := 0; ri < DIM; ri++ {
		for ci := 0; ci < DIM; ci++ {
			if problem[ri][ci] != 0 {
				continue
			}
			bitCnt := bc(candidates[ri][ci])
			if bitCnt < count {
				count, r, c = bitCnt, ri, ci
			}
		}
	}

	var status Status
	switch count {
	case DIM + 1: // no empty cell left
		status = solved
	case 0: // empty cell with no candidate
		status = unsolvable
	default: // empty cell with at least 1 candidate
		status = notSolvedYet
	}
	return Cell{r, c}, status
}

type Unit [9]Cell
type UContext struct {
	pos   []Cell
	count int
	value uint8
}

func processUnit(ctx *UContext, b *Board, candidates *Candidates, unit Unit) {
	var cnt [9]int
	var bestValue uint8 = DIM + 1
	var missing uint16 = ALL

	for i := range unit {
		cellR, cellC := unit[i].r, unit[i].c
		v := b[cellR][cellC]
		if v != 0 {
			missing &= ^singleton(v)
		} else {
			cand := candidates[cellR][cellC]
			for x := 0; x < DIM; x++ {
				if (cand & singleton(uint8(x+1))) != 0 {
					cnt[x]++
				}
			}
		}
	}

	// Find the missing value with the fewest available slots
	for v := 0; v < DIM; v++ {
		if (missing&singleton(uint8(v+1))) != 0 && (bestValue == DIM+1 || cnt[v] < cnt[bestValue]) {
			bestValue = uint8(v)
		}
	}

	if bestValue == DIM+1 { // Find nothing
		return
	}

	if ctx.count == -1 || cnt[bestValue] < ctx.count {
		ctx.value = bestValue + 1
		ctx.count = cnt[bestValue]
		mask := singleton(bestValue + 1)
		ctx.pos = make([]Cell, 0, ctx.count)
		for i := range unit {
			if (b[unit[i].r][unit[i].c] == 0) && ((candidates[unit[i].r][unit[i].c] & mask) != 0) {
				ctx.pos = append(ctx.pos, unit[i])
			}
		}
	}
}

func fewestCells(b *Board, candidates *Candidates) *UContext {
	ctx := &UContext{
		count: -1,
		value: DIM + 1,
		pos:   nil,
	}

	// Row units
	for ri := 0; ri < DIM; ri++ {
		var unit Unit
		for ci := 0; ci < DIM; ci++ {
			unit[ci] = Cell{ri, ci}
		}
		processUnit(ctx, b, candidates, unit)
	}
	// Column units
	for ci := 0; ci < DIM; ci++ {
		var unit Unit
		for ri := 0; ri < DIM; ri++ {
			unit[ri] = Cell{ri, ci}
		}
		processUnit(ctx, b, candidates, unit)
	}
	// Box units
	for ri := 0; ri < DIM/BOX; ri++ {
		for ci := 0; ci < DIM/BOX; ci++ {
			var unit Unit
			var id int
			for i := 0; i < BOX; i++ {
				for j := 0; j < BOX; j++ {
					unit[id] = Cell{ri*BOX + i, ci*BOX + j}
					id++
				}
			}
			if id != DIM {
				log.Fatalf("fewestCells: wrong box unit size %d", id)
			}
			processUnit(ctx, b, candidates, unit)
		}
	}
	return ctx
}

type SContext struct {
	problem, solution    *Board
	nSolution, diffScore int
}

func recSolve(ctx *SContext, b *Board, candidates *Candidates, score int) {
	cell, status := fewestCandidates(b, candidates)
	if status == solved {
		if ctx.nSolution == 0 {
			ctx.diffScore = score
			ctx.solution = b.copy()
		}
		ctx.nSolution++
		return
	}

	if status == unsolvable {
		ctx.nSolution = 0
		ctx.diffScore = -1 // invalid score
		ctx.solution = nil
		return
	}

	mask := candidates[cell.r][cell.c]
	if (mask & (mask - 1)) > 0 { // more than one candidate -> try fewestCells strategy
		uc := fewestCells(b, candidates)
		if uc.count > 0 && uc.count < int(bc(mask)) {
			bf := uc.count - 1
			score += bf * bf
			for i := 0; i < uc.count; i++ {
				tempCand := candidates.copy()
				// try to put uc.value at (pos[i].r, pos[i].c)
				tempCand.eliminate(uc.pos[i].r, uc.pos[i].c, uc.value)
				b[uc.pos[i].r][uc.pos[i].c] = uc.value
				recSolve(ctx, b, tempCand, score)
				b[uc.pos[i].r][uc.pos[i].c] = 0
				if ctx.nSolution >= 2 {
					return
				}
			}
			return
		}
	}

	bf := int(bc(mask)) - 1
	score += bf * bf
	for v := 0; v < DIM; v++ {
		if (mask & singleton(uint8(v+1))) != 0 {
			tempCand := candidates.copy()
			tempCand.eliminate(cell.r, cell.c, uint8(v+1))
			b[cell.r][cell.c] = uint8(v + 1)
			recSolve(ctx, b, tempCand, score)
			if ctx.nSolution >= 2 {
				return
			}
		}
	}
	b[cell.r][cell.c] = 0
}

func Solve(input *Board) (*SContext, error) {
	ctx := &SContext{
		nSolution: 0,
		diffScore: -1,
		problem:   input.copy(),
		solution:  nil,
	}

	candidates := input.setCandidates()
	if !isValid(input, candidates) {
		return nil, ErrInvalidInput
	}
	var score int
	recSolve(ctx, input, candidates, score)
	if ctx.diffScore > 0 {
		var nEmpty int
		for ri := 0; ri < DIM; ri++ {
			for ci := 0; ci < DIM; ci++ {
				if ctx.problem[ri][ci] == 0 {
					nEmpty++
				}
			}
		}
		ctx.diffScore = ctx.diffScore*100 + nEmpty
	}
	switch ctx.nSolution {
	case 0:
		return ctx, ErrNoSolution
	case 1:
		return ctx, nil
	default:
		return ctx, ErrMultipleSolution
	}
}

func diffLevel(diffScore int) string {
	if diffScore <= 250 {
		return "Easy"
	} else if diffScore <= 500 {
		return "Medium"
	}
	return "Hard"

}

func printSolution(b *Board, score int) {
	fmt.Printf("Solution:\n%v", b)
	fmt.Printf("Difficulty level: %s\n", diffLevel(score))
}

func ReportSolution(ctx *SContext, err error) {
	switch err {
	case ErrInvalidInput:
		fmt.Printf("Invalid input:\n%v\n", ctx.problem)
	case ErrNoSolution:
		fmt.Printf("No solution found for:\n%v\n", ctx.problem)
	case ErrMultipleSolution:
		fmt.Printf("Multiple solutions exist for:\n%v\n", ctx.problem)
		printSolution(ctx.solution, ctx.diffScore)
	default:
		fmt.Printf("Problem:\n%v\n", ctx.problem)
		printSolution(ctx.solution, ctx.diffScore)
	}
}

func LoadInput(reader io.Reader) (*Board, error) {
	scanner := bufio.NewScanner(reader)
	input := new(Board)
	var row int
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != DIM {
			return nil, ErrMalformedInput
		}
		for i, v := range fields {
			switch v {
			case "_":
				input[row][i] = uint8(0)
			default:
				val, err := strconv.ParseUint(v, 10, 8)
				if err != nil {
					log.Println("fail to parse input cell; expected a digit or _, but got %v: %v", v, err)
					return nil, ErrMalformedInput
				}
				input[row][i] = uint8(val)
			}
		}
		row++
	}
	if row != DIM {
		return nil, ErrMalformedInput
	}
	if err := scanner.Err(); err != nil {
		return nil, ErrMalformedInput
	}
	return input, nil
}
