package main

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func TestEmptyRects(t *testing.T) {
	// arrange
	board := `400103000080507420900040005139000500270910040804730912592080000748350290000279854`

	b, err := loadBoard([]byte(board))
	if err != nil {
		t.Fatal(err)
	}

	//b.PrintHints()

	// check board is in expected initial state
	testHint(t, b, 3, 8, []uint{6, 7, 8})

	// act
	b.changed = false
	if err = b.SolveEmptyRectangles(); err != nil {
		t.Fatal(err)
	}

	// assert
	if !b.changed {
		t.Fatal("board not changed")
	}

	//b.PrintHints()

	testHint(t, b, 3, 8, []uint{7, 8})
}

func TestEmptyRects2(t *testing.T) {
	// arrange
	hintBoard := `
|---|-------------------------------------------------|-------------------------------------------------|-------------------------------------------------|
|r,c|               0               1               2 |               3               4               5 |               6               7               8 |
|---|-------------------------------------------------|-------------------------------------------------|-------------------------------------------------|
| 0 |           (2,5)       (1,2,6,7)               8 |               4         (2,5,7)       (2,5,6,9) |       (1,2,6,9)               3           (1,9) |
| 1 |         (2,4,5)       (1,2,6,7)         (1,4,7) |               3         (2,5,7)       (2,5,6,9) |     (1,2,6,8,9)       (1,6,8,9)         (1,8,9) |
| 2 |               9         (2,3,6)         (2,3,6) |         (2,6,8)           (2,8)               1 |               5               7               4 |
|---|-------------------------------------------------|-------------------------------------------------|-------------------------------------------------|
| 3 |               7               9       (2,3,5,6) |       (1,2,5,6)     (1,2,3,4,5)               8 |         (1,3,6)         (1,4,6)           (1,3) |
| 4 |         (2,3,8)       (2,3,6,8)         (2,3,6) |         (1,2,6)       (1,2,3,4)               7 |       (1,3,6,9)       (1,4,6,9)               5 |
| 5 |               1               4         (3,5,6) |               9           (3,5)         (3,5,6) |           (7,8)               2           (7,8) |
|---|-------------------------------------------------|-------------------------------------------------|-------------------------------------------------|
| 6 |         (3,4,8)       (1,3,7,8)               9 |       (1,5,7,8)               6         (3,4,5) |       (1,3,7,8)           (1,8)               2 |
| 7 |               6               5           (1,7) |       (1,2,7,8)       (1,2,3,8)           (2,3) |               4         (1,8,9)     (1,3,7,8,9) |
| 8 |       (2,3,4,8)     (1,2,3,7,8)         (1,4,7) |         (1,7,8)               9           (3,4) |       (1,3,7,8)               5               6 |
|---|-------------------------------------------------|-------------------------------------------------|-------------------------------------------------|
`

	b := loadBoardWithHints(t, hintBoard)

	// check board is in expected initial state
	testHint(t, b, 4, 1, []uint{2, 3, 6, 8})

	// act
	b.changed = false
	if err := b.SolveEmptyRectangles(); err != nil {
		t.Fatal(err)
	}

	// assert
	testHint(t, b, 4, 1, []uint{2, 3, 6, 8})

	if b.changed {
		t.Fatal("board changed, no empty-rectangle options here")
	}
}

func TestEmptyRects3(t *testing.T) {
	// arrange
	board := `750960320000702050000030047970050083005070100180000075240090710010407000097016030`

	b, err := loadBoard([]byte(board))
	if err != nil {
		t.Fatal(err)
	}

	if err = b.SolveWithSolversList(b.getSimpleSolvers()); err != nil {
		t.Fatal(err)
	}

	// check board is in expected initial state
	testHint(t, b, 4, 8, []uint{2, 4, 6, 9})
	testHint(t, b, 7, 2, []uint{3, 6, 8})

	// act
	if err = b.SolveEmptyRectangles(); err != nil {
		t.Fatal(err)
	}

	// assert
	if !b.changed {
		t.Fatal("board not changed")
	}
}

func TestXYChain(t *testing.T) {
	// arrange
	hintBoard := `
|---|-------------------------------------------------|-------------------------------------------------|-------------------------------------------------|
|r,c|               0               1               2 |               3               4               5 |               6               7               8 |
|---|-------------------------------------------------|-------------------------------------------------|-------------------------------------------------|
| 0 |               4               8               7 |               3               1               2 |           (5,6)               9           (5,6) |
| 1 |           (5,9)           (5,9)               3 |               6           (4,8)           (4,8) |               2               7               1 |
| 2 |               1               2               6 |           (5,7)               9           (5,7) |               3               8               4 |
|---|-------------------------------------------------|-------------------------------------------------|-------------------------------------------------|
| 3 |               7           (3,4)               5 |           (8,9)         (3,4,8)         (4,8,9) |               1               6               2 |
| 4 |           (6,9)               1           (4,9) |               2         (3,4,6)           (5,7) |               8           (3,4)           (5,7) |
| 5 |           (2,8)         (3,4,6)           (2,8) |           (5,7)         (3,4,6)               1 |           (5,7)           (3,4)               9 |
|---|-------------------------------------------------|-------------------------------------------------|-------------------------------------------------|
| 6 |           (5,8)           (4,5)               1 |           (4,8)               7               6 |               9               2               3 |
| 7 |               3           (6,7)           (8,9) |               1               2           (8,9) |               4               5           (6,7) |
| 8 |           (2,6)       (4,6,7,9)         (2,4,9) |           (4,9)               5               3 |           (6,7)               1               8 |
|---|-------------------------------------------------|-------------------------------------------------|-------------------------------------------------|
	`
	b := loadBoardWithHints(t, hintBoard)

	// check board is in expected initial state
	testHint(t, b, 5, 6, []uint{5, 7})
	testHint(t, b, 3, 1, []uint{3, 4})
	testHint(t, b, 5, 1, []uint{3, 4, 6})
	testHint(t, b, 8, 2, []uint{2, 4, 9})

	// act
	if err := b.SolveXYChain(); err != nil {
		t.Fatal(err)
	}

	// assert

	// test for absence of defect in XY-Chain:
	// - R5C6: old hints: 5,7        remove hint: 5 remaining hints: 7
	testHint(t, b, 5, 6, []uint{5, 7})

	// test for expected state after XY-Chain applied
	// note: this test may fail in the future if XY-Chain is modified
	//       since it could pick a different chain to operate on
	// - R3C1: old hints: 3,4        remove hint: 4 remaining hints: 3
	// - R5C1: old hints: 3,4,6      remove hint: 4 remaining hints: 3,6
	// - R8C2: old hints: 2,4,9      remove hint: 4 remaining hints: 2,9
	testHint(t, b, 3, 1, []uint{3})
	testHint(t, b, 5, 1, []uint{3, 6})
	testHint(t, b, 8, 2, []uint{2, 9})
}

func testHint(t *testing.T, b *board, row, col int, hints []uint) {
	actual := b.blits[row*9+col]
	var expected uint
	for _, hint := range hints {
		expected |= 1 << (hint - 1)
	}
	if expected != actual {
		t.Fatalf("R%dC%d, expected %v actual %v", row, col, hints, GetBitsString(actual))
	}
}

func loadBoardWithHints(t *testing.T, hintBoard string) (b *board) {
	// read the text board, apply hints
	var err error
	sr := strings.NewReader(hintBoard)
	r := bufio.NewReader(sr)

	// skip header
	for i := 0; i < 4; i++ {
		if _, err = r.ReadString('\n'); err != nil {
			t.Fatal(err)
		}
	}

	b = &board{}
	var line string
	for i := 0; i < 9; i++ {
		if line, err = r.ReadString('\n'); err != nil {
			t.Fatal(err)
		}

		if strings.HasPrefix(line, "|---|") {
			if line, err = r.ReadString('\n'); err != nil {
				t.Fatal(err)
			}
		}

		line = strings.Replace(line, "\r", "", -1)
		line = strings.Replace(line, "\n", "", -1)
		line = line[6 : len(line)-2]
		line = strings.Replace(line, " |", "", -1)

		start := 0
		cells := make([]string, 9)
		for j := 0; j < 9; j++ {
			end := start + 15
			cell := strings.Trim(line[start:end], " ")
			cells[j] = cell
			start = end + 1

			pos := i*9 + j
			if strings.HasPrefix(cell, "(") {
				// get hints
				hints := strings.Split(cell[1:len(cell)-1], ",")
				for _, hint := range hints {
					val, err := strconv.Atoi(hint)
					if err != nil {
						t.Fatal(err)
					}
					b.blits[pos] |= 1 << uint(val-1)
				}
			} else {
				// solved cell
				val, err := strconv.Atoi(cell)
				if err != nil {
					t.Fatal(err)
				}
				b.solved[pos] = uint(val)
				b.blits[pos] = 1 << uint(val-1)
			}
		}
	}

	return b
}

func TestBoards(t *testing.T) {
	files := []string{
		"./test_files/input.txt",
		"./test_files/01_naked_single_493382.txt",
		"./test_files/02_hidden_single_1053217.txt",
		"./test_files/03_naked_pair_1053222.txt",
		"./test_files/04_naked_triple_1043003.txt",
		"./test_files/05_naked_quint_1051073.txt",
		"./test_files/06_hidden_pair_1208057.txt",
		"./test_files/07_hidden_triple_188899.txt",
		"./test_files/08_hidden_quint_188899.txt",
		"./test_files/09_pointing_pair_and_triple_1011509.txt",
		"./test_files/10_xwing_1307267.txt",
		"./test_files/12_tough_20151107_173.txt",
		"./test_files/11_swordfish_1280430.txt",
		"./test_files/13_swordfish_008009000300057001000100009230000070005406100060000038900003000700840003000700600.txt",
		"./test_files/14_swordfish_980010020002700000000009010700040800600107002009030005040900000000005700070020039.txt",
		"./test_files/15_swordfish_108000067000050000000000030006100040450000900000093000200040010003002700807001005.txt",
		"./test_files/16_swordfish_107300040800006000050870630090000510000000007700060080000904000080100002410000000.txt",
		"./test_files/17_swordfish_300040000000007048000000907010003080400050020050008070500300000000000090609025300.txt",
		"./test_files/18_swordfish.txt",
		"./test_files/19_supposedly_hard.txt",
		"./test_files/20_17_clues.txt",
		"./test_files/21_ywing.txt",
		"./test_files/22_xychain.txt",
		"./test_files/23_xychain.txt",
		"./test_files/24_xychain.txt",
		"./test_files/25_xychain.txt",
		"./test_files/26_xychain.txt",
		"./test_files/27_xcycles.txt",
		"./test_files/28_xcycles.txt",
		"./test_files/29_ben.txt",
		"./test_files/30_starburst_leo.txt",
	}

	for _, file := range files {
		board, err := getBoard(file)
		if err != nil {
			t.Fatalf("%s: %s", file, err)
			return
		}

		if err = board.Solve(); err != nil {
			board.PrintHints()
			t.Fatalf("%s: %s", file, err)
			return
		}

		if !board.isSolved() {
			board.PrintHints()
			board.PrintCompact()
			t.Fatalf("%s: could not solve", file)
			return
		}
	}
	fmt.Printf("solved %d puzzles\n", len(files))
}

func getKnownAnswer(t *testing.T, answer string) *[81]byte {
	if len(answer) != 81 {
		t.Errorf("len(answer) == %d, expected 81", len(answer))
		return nil
	}

	var ka [81]byte
	for i := 0; i < 81; i++ {
		ka[i] = answer[i] - 48
	}
	return &ka
}

func Test13329(t *testing.T) {
	// arrange
	hintBoard := `
|---|-------------------------------------------------|-------------------------------------------------|-------------------------------------------------|
|r,c|               0               1               2 |               3               4               5 |               6               7               8 |
|---|-------------------------------------------------|-------------------------------------------------|-------------------------------------------------|
| 0 |       (1,5,8,9)         (1,7,9)               2 |               6       (4,5,8,9)     (1,4,5,8,9) |       (5,7,8,9)               3     (1,4,5,8,9) |
| 1 |               4       (1,3,6,9)     (1,5,6,8,9) |       (1,2,5,9)               7     (1,3,5,8,9) |     (2,5,6,8,9)     (1,2,5,8,9)   (1,2,5,6,8,9) |
| 2 |   (1,3,5,6,8,9)     (1,3,6,7,9)   (1,5,6,7,8,9) |     (1,2,4,5,9)   (2,3,4,5,8,9)   (1,3,4,5,8,9) |   (2,5,6,7,8,9) (1,2,4,5,7,8,9) (1,2,4,5,6,8,9) |
|---|-------------------------------------------------|-------------------------------------------------|-------------------------------------------------|
| 3 |       (2,5,6,9)         (2,6,9)         (5,6,9) |         (2,5,9)               1     (3,5,6,8,9) |               4       (2,5,8,9)               7 |
| 4 |     (1,2,5,6,9)               8               3 |     (2,4,5,7,9)     (2,4,5,6,9)     (4,5,6,7,9) |       (2,5,6,9)         (2,5,9)       (2,5,6,9) |
| 5 |       (2,5,6,9)     (2,4,6,7,9)     (4,5,6,7,9) |     (2,4,5,7,9) (2,3,4,5,6,8,9) (3,4,5,6,7,8,9) |               1       (2,5,8,9)   (2,3,5,6,8,9) |
|---|-------------------------------------------------|-------------------------------------------------|-------------------------------------------------|
| 6 |         (1,8,9)         (1,4,9)       (1,4,8,9) |               3         (4,5,9)               2 |       (5,7,8,9)               6     (1,4,5,8,9) |
| 7 |               7               5     (1,4,6,8,9) |         (1,4,9)         (4,6,9)       (1,4,6,9) |       (2,3,8,9)     (1,2,4,8,9)   (1,2,3,4,8,9) |
| 8 |     (1,2,3,6,9)   (1,2,3,4,6,9)       (1,4,6,9) |               8       (4,5,6,9)   (1,4,5,6,7,9) |     (2,3,5,7,9)   (1,2,4,5,7,9)   (1,2,3,4,5,9) |
|---|-------------------------------------------------|-------------------------------------------------|-------------------------------------------------|
`

	b := loadBoardWithHints(t, hintBoard)
	//b.PrintURL()
	//b.knownAnswer = getKnownAnswer(t, "872691534431275896695438271269513487183724659547986123914352768758169342326847915")

	/*
	   872 691 534
	   431 275 896
	   695 438 271

	   269 513 487
	   183 724 659
	   547 986 123

	   914 352 768
	   758 169 342
	   326 847 915
	*/

	//b.quiet = true
	//b.verbose = true

	// check board is in expected initial state
	//testHint(t, b, 5, 6, []uint{5, 7})
	//testHint(t, b, 3, 1, []uint{3, 4})
	//testHint(t, b, 5, 1, []uint{3, 4, 6})
	//testHint(t, b, 8, 2, []uint{2, 4, 9})

	// act / assert
	if err := b.Solve(); err != nil {
		t.Fatal(err)
	}
}

func TestXWing(t *testing.T) {
	// arrange
	hintBoard := `
|---|-------------------------------------------------|-------------------------------------------------|-------------------------------------------------|
|r,c|               0               1               2 |               3               4               5 |               6               7               8 |
|---|-------------------------------------------------|-------------------------------------------------|-------------------------------------------------|
| 0 |               8               7               2 |               6           (4,9)         (1,5,9) |           (5,9)               3       (1,4,5,9) |
| 1 |               4               3           (1,5) |           (1,2)               7       (1,5,8,9) |           (2,8)         (1,5,9)               6 |
| 2 |               6               9           (1,5) |         (1,2,4)               3         (1,5,8) |           (2,8)               7         (1,4,5) |
|---|-------------------------------------------------|-------------------------------------------------|-------------------------------------------------|
| 3 |               2               6               9 |               5               1               3 |               4               8               7 |
| 4 |               1               8               3 |               7               2               4 |               6           (5,9)           (5,9) |
| 5 |               5               4               7 |               9               8               6 |               1               2               3 |
|---|-------------------------------------------------|-------------------------------------------------|-------------------------------------------------|
| 6 |               9               1               4 |               3               5               2 |               7               6               8 |
| 7 |               7               5               8 |           (1,4)               6           (1,9) |               3           (4,9)               2 |
| 8 |               3               2               6 |               8           (4,9)               7 |           (5,9)       (1,4,5,9)         (1,5,9) |
|---|-------------------------------------------------|-------------------------------------------------|-------------------------------------------------|
	`
	b := loadBoardWithHints(t, hintBoard)
	//b.PrintHints()
	//b.PrintURL()
	//b.knownAnswer = getKnownAnswer(t, "872691534431275896695438271269513487183724659547986123914352768758169342326847915")

	/*
	   872 691 534
	   431 275 896
	   695 438 271

	   269 513 487
	   183 724 659
	   547 986 123

	   914 352 768
	   758 169 342
	   326 847 915
	*/

	//b.quiet = true
	//b.verbose = true

	// act / assert
	if err := b.SolveXWing(); err != nil {
		b.PrintHints()
		t.Fatal(err)
	}

	// try it again
	if err := b.SolveXWing(); err != nil {
		b.PrintHints()
		t.Fatal(err)
	}
}

func (b *board) getSimpleSolvers() []solver {
	solvers := []solver{
		{name: "NAKED SINGLE", run: b.SolveNakedSingle},
		{name: "HIDDEN SINGLE", run: b.SolveHiddenSingle},
		{name: "NAKED PAIR", run: b.getSolverN(b.SolveNakedN, 2)},
		{name: "NAKED TRIPLE", run: b.getSolverN(b.SolveNakedN, 3)},
		{name: "NAKED QUAD", run: b.getSolverN(b.SolveNakedN, 4)},
		{name: "NAKED QUINT", run: b.getSolverN(b.SolveNakedN, 5)},
		{name: "HIDDEN PAIR", run: b.getSolverN(b.SolveHiddenN, 2)},
		{name: "HIDDEN TRIPLE", run: b.getSolverN(b.SolveHiddenN, 3)},
		{name: "HIDDEN QUAD", run: b.getSolverN(b.SolveHiddenN, 4)},
		{name: "HIDDEN QUINT", run: b.getSolverN(b.SolveHiddenN, 5)},
		{name: "POINTING PAIR AND TRIPLE REDUCTION", run: b.SolvePointingPairAndTripleReduction},
		{name: "BOX LINE", run: b.SolveBoxLine},
	}

	return solvers
}
