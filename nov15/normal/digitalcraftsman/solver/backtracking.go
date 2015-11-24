// Package solver implements methods to solve Sudoku games
// automatically using the backtracking algorithm.
package solver

// Backtrack tries to find a valid solution for a given Board.
// A returned Boolean indicates if the board was solved successfully.
func (b *Board) Backtrack() bool {
	nextRow, nextCol, hasEmptyCell := b.findEmptyCell()
	if !hasEmptyCell {
		return true
	}

	for candidate := 1; candidate <= N; candidate++ {
		if b.isDigitValid(nextRow, nextCol, candidate) {
			b.Cells[nextRow][nextCol] = candidate

			if b.Backtrack() {
				return true
			}
			// reset the cell
			b.Cells[nextRow][nextCol] = 0
		}
	}

	return false
}

// findEmptyCell checks if an empty cell exists. Returned are the row and
// column of the empty cell and an indicator of their existence. If no
// empty cell exists the row and column will default to 0.
func (b *Board) findEmptyCell() (int, int, bool) {
	for row := 0; row < N; row++ {
		for col := 0; col < N; col++ {
			if b.Cells[row][col] == 0 {
				return row, col, true
			}
		}
	}

	return 0, 0, false
}

// isDigitValid checks wether a given digit already appears
// in the corresponding row, column or 3x3 section.
func (b *Board) isDigitValid(row, col, digit int) bool {
	// note: integer division 'rounds down' by ignoring all decimal places
	startRow := row / 3 * 3
	startCol := col / 3 * 3

	for i := 0; i < N; i++ {
		// check the corresponding row and column
		if b.Cells[row][i] == digit ||
			b.Cells[i][col] == digit ||
			// check the corresponding 3x3 section
			b.Cells[startRow+i/3][startCol+i%3] == digit {
			return false
		}
	}

	return true
}
