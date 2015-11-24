package main

import "sync"

// Solve solves the sudoku puzzle given by b. The return value indicates
// whether a solution could be found. If there are multiple solutions to the
// puzzle, the first one found by the search algorithm will be used and the
// others are discarded.
func (b *Board) Solve() bool {

	err := b.solveFixpoint()
	if err != nil {
		// if solving to fixed point reveals an error, don't attempt to solve further
		return false
	}

	if b.isSolved() {
		return true
	}

	solved := make(chan Board)
	quit := make(chan struct{})
	done := make(chan struct{})
	var wg sync.WaitGroup

	wg.Add(1)
	nb := *b
	go nb.solveConcurrent(solved, quit, &wg)

	go func() {
		wg.Wait()
		close(done)
	}()

	// either we get a solved board, or all calls to solveConcurrent finish
	// without finding a solution
	select {
	case *b = <-solved:
	case <-done:
	}

	close(quit)
	return b.isSolved()
}

// solveConcurrent takes the given board and does elimination to fixed point.
// If this does not solve the board completely, it will take a space with
// multiple possible values and spawn a new goroutine calling solveConcurrent
// for each possible value.
// The quit channel is used to terminate all calls to solveConcurrent as soon
// as a solution is found and prevent goroutine leaks.
// The sync.WaitGroup is used to indicate to Solve() when all instances of
// solveConcurrent have terminated. This is necessary in the case where
// there is no solution, so Solve() does not wait on the solved channel forever.
func (b *Board) solveConcurrent(solved chan<- Board, quit <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	select {
	case <-quit:
		return
	default:
	}

	err := b.solveFixpoint()
	if err != nil {
		return
	}

	pos, s, err := b.findNonfixed()
	if err != nil { // no nonfixed squares, then b is solved
		select {
		case solved <- *b:
		case <-quit:
		}
		return
	}

	for v, p := range s.possible {
		if p {
			nb := *b
			nb[pos.row][pos.col].Set(v)
			wg.Add(1)
			go nb.solveConcurrent(solved, quit, wg)
		}
	}
}

// solveFixpoint does the following until no more changes can be made to
// the board: for each solved space on the board, prune that space's value from
// all peers of that space. This occurs only once for each solved space.
// (This is achieved by setting the field usedForPrune in the space.)
// A non-nil error is returned if one of the required pruning operations is
// impossible, indicating that the board has no solution. When returning
// no error, the board is either completely solved or in a valid state with
// some spaces not yet solved.
func (b *Board) solveFixpoint() error {
	for {
		didWork := false
		for _, pv := range b.getUnusedFixed() {
			for _, s := range b.getPeers(pv.pos) {
				work, err := s.prune(pv.value)
				if err != nil {
					return err
				}
				if work {
					didWork = true
				}
			}
		}
		if !didWork {
			break
		}
	}
	return nil
}

func (b *Board) isSolved() bool {
	_, _, err := b.findNonfixed()
	return err != nil
}
