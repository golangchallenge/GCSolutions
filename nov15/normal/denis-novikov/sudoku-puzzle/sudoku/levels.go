package sudoku

// DifficultyLevel is a difficulty level of a puzzle.
type DifficultyLevel int

const (
	// DLUnknown is returned when an error is returned or when a puzzle is
	// not solved.
	DLUnknown DifficultyLevel = iota
	// DLEasy is an easy level for sudoku a puzzle.
	DLEasy
	// DLMedium ia a medium level for a sudoku puzzle.
	DLMedium
	// DLHard is a hard level for a sudoku puzzle.
	DLHard
)

// String converts a puzzle to string.
func (dl DifficultyLevel) String() string {
	if dl == DLUnknown {
		return "Unknown"
	}
	if dl == DLEasy {
		return "Easy"
	}
	if dl == DLMedium {
		return "Medium"
	}
	return "Hard"
}

func (p *Puzzle) getDifficulty() DifficultyLevel {
	if p.err != nil || p.loopCount == 0 {
		return DLUnknown
	}

	tgDiff := uint(0)
	if p.emptyCells > 50 {
		tgDiff++
		if p.emptyCells > 53 {
			tgDiff++
			if p.emptyCells > 57 {
				tgDiff++
				if p.emptyCells > 60 {
					tgDiff++
				}
			}
		}
	}

	lbDiff := uint(0)
	if p.lbRow == 0 {
		lbDiff += 4
	} else if p.lbRow < 5 {
		lbDiff += 5 - uint(p.lbRow)
	}
	if p.lbCol == 0 {
		lbDiff += 4
	} else if p.lbCol < 5 {
		lbDiff += 5 - uint(p.lbCol)
	}
	if p.lbBlock == 0 {
		lbDiff += 4
	} else if p.lbBlock < 5 {
		lbDiff += 5 - uint(p.lbBlock)
	}

	lbDiff /= 3

	lcDiff := uint(0)

	if p.loopCount > 600 {
		lcDiff++
		if p.loopCount > 750 {
			lcDiff++
			if p.loopCount > 1000 {
				lcDiff++
				if p.loopCount > 1500 {
					lcDiff++
				}
			}
		}
	}

	//difficulty := ((tgDiff * 6) + lbDiff + lcDiff*3) / 12
	difficulty := (tgDiff*2 + lbDiff + lcDiff*5) / 9

	if difficulty < 2 {
		return DLEasy
	}
	if difficulty == 2 {
		return DLMedium
	}
	return DLHard
}
