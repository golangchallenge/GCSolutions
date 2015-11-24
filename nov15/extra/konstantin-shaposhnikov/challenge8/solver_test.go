package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

type testcase struct {
	puzzle     Grid
	solution   Grid
	skipReason string
	level      int
}

func readTestcase(path string) (*testcase, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read test %s: %s", path, err)
	}
	defer f.Close()

	var t testcase

	r := bufio.NewReader(f)
	grids := 0
ParseLoop:
	for {
		var line string
		if line, err = r.ReadString('\n'); err != nil {
			break
		}
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "//"):
			continue // skip comments
		case line == "":
			continue // skip empty lines
		case strings.HasPrefix(line, "Skip:"):
			t.skipReason = line
			break ParseLoop
		case strings.HasPrefix(line, "Level:"):
			s := strings.TrimSpace(strings.TrimPrefix(line, "Level:"))
			if t.level, err = strconv.Atoi(s); err != nil {
				break ParseLoop
			}
			continue
		case line == "Puzzle:":
			if t.puzzle, err = ReadGrid(r); err != nil {
				break ParseLoop
			}
			grids++
		case line == "Solution:":
			if t.solution, err = ReadGrid(r); err != nil {
				break ParseLoop
			}
			grids++
		default:
			err = fmt.Errorf("unexpected line: %s", line)
			break ParseLoop
		}
	}
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read test case from %s: %s", path, err)
	}
	if t.skipReason == "" && grids != 2 {
		return nil, fmt.Errorf("invalid test case file %s (no puzzle and/or solution grid found)", path)
	}

	return &t, nil
}

func findTestcases() []string {
	fileNames, err := ioutil.ReadDir("testdata")
	if err != nil {
		panic(fmt.Sprintf("Failed to read tests: %s", err))
	}

	var filePaths []string
	for _, test := range fileNames {
		filePaths = append(filePaths, filepath.Join("testdata", test.Name()))
	}
	return filePaths
}

func TestPuzzles(t *testing.T) {
	for _, p := range findTestcases() {
		test, err := readTestcase(p)
		if err != nil {
			t.Errorf("failed to read test %s: %s", p, err)
			continue
		}

		if test.skipReason != "" {
			t.Logf("skipping test case %s: %s", p, test.skipReason)
			continue
		}

		// solve and compare answer
		s := newSolver(test.puzzle)
		s.solve()
		if s.err != nil {
			t.Errorf("ubable to solve puzzle (%s): %s\n%s\ncould only get so far:\n%s",
				p, s.err, test.puzzle, s.g)
			continue
		}
		if s.g != test.solution {
			t.Errorf("failed to solve puzzle (%s):\n%s\nGot:\n%s\nWant:\n%s\n",
				p, test.puzzle, s.g, test.solution)
			continue
		}
		if s.level() != test.level {
			t.Errorf("failed to estimate puzzle (%s) difficulty: got %d, want: %d",
				p, s.level(), test.level)
			continue
		}
		t.Logf("solved puzzle from %s (Level: %d)", p, test.level)
	}
}

func BenchmarkP096(b *testing.B) {
	var tests []*testcase
	for _, p := range findTestcases() {
		if strings.Index(p, "p096_") >= 0 {
			test, err := readTestcase(p)
			if err != nil {
				panic(err)
			}
			tests = append(tests, test)
		}
	}
	if len(tests) != 50 {
		panic("expect 50 tests")
	}

	for i := 0; i < b.N; i++ {
		for _, test := range tests {
			s := newSolver(test.puzzle)
			s.solve()
			if s.err != nil {
				panic(s.err)
			}
		}
	}
}

func setCandidates(g group, candidates [GridSize][]byte) {
	for i := 0; i < GridSize; i++ {
		g.cell(i).set(candidates[i])
	}
}

func candidates(g group) [GridSize][]byte {
	var candidates [GridSize][]byte
	for i := 0; i < GridSize; i++ {
		candidates[i] = g.cell(i).get()
	}
	return candidates
}

func TestRegionLockedCandidates1(t *testing.T) {
	g := NewGrid()
	// this region has only candidate 2's in its bottom row
	// and only candidate 1's in its most right column
	region := g.region(GridSize - 1)
	setCandidates(region, [GridSize][]byte{
		{4},
		{3},
		{1, 5, 6, 8},
		{9},
		{5, 8},
		{1, 5, 8},
		{2, 6, 8},
		{2, 5, 6, 8},
		{7},
	})
	// Therefore 2 can be excluded as a candidate from the marked cells.
	row := g.row(GridSize - 1)
	setCandidates(row, [GridSize][]byte{
		{1, 4, 6},
		{9},
		{1, 4, 5},
		{2, 4, 5},       // 2 cannot be here
		{1, 2, 4, 5, 8}, // 2 cannot be here
		{3},
		{2, 6, 8},
		{2, 5, 6, 8},
		{7},
	})
	// And 1 can be excluded as a candidate from the marked cells.
	col := g.column(GridSize - 1)
	setCandidates(col, [GridSize][]byte{
		{9},
		{2},
		{1, 3, 4}, // 1 cannot be here
		{1, 5, 8}, // 1 cannot be here
		{4, 5, 6},
		{1, 3, 4, 8}, // 1 cannot be here
		{1, 5, 6, 8},
		{1, 5, 8},
		{7},
	})

	regionLockedCandidates1(&g, GridSize-1)

	wantRow := [GridSize][]byte{
		{1, 4, 6},
		{9},
		{1, 4, 5},
		{4, 5},
		{1, 4, 5, 8},
		{3},
		{2, 6, 8},
		{2, 5, 6, 8},
		{7},
	}
	if got := candidates(row); !reflect.DeepEqual(wantRow, got) {
		t.Errorf("want (row): %v, got: %v", wantRow, got)
	}

	wantCol := [GridSize][]byte{
		{9},
		{2},
		{3, 4},
		{5, 8},
		{4, 5, 6},
		{3, 4, 8},
		{1, 5, 6, 8},
		{1, 5, 8},
		{7},
	}
	if got := candidates(col); !reflect.DeepEqual(wantCol, got) {
		t.Errorf("want (column): %v, got: %v", wantCol, got)
	}
}

func TestRegionLockedCandidates2(t *testing.T) {
	g := NewGrid()

	// This col (#6) contains 9 only in region 5
	col6 := g.column(6)
	setCandidates(col6, [GridSize][]byte{
		{4},
		{2, 3, 5},
		{6},
		{5, 7, 9},
		{2, 7},
		{2, 5, 9},
		{1},
		{3, 5, 7},
		{8},
	})

	// So all 9's can be excluded from cells in column 7 and 8
	rgn5 := g.region(5)
	setCandidates(rgn5, [GridSize][]byte{
		{5, 7, 9},
		{1, 4, 5, 7, 9}, // 9 cannot be here
		{1, 4, 5, 6, 9}, // 9 cannot be here
		{2, 7},
		{3},
		{1, 8},
		{2, 5, 9},
		{2, 5, 9},    // 9 cannot be here
		{5, 6, 8, 9}, // 9 cannot be here
	})

	rowColLockedCandidates2(&g, col6)

	wantRgn := [GridSize][]byte{
		{5, 7, 9},
		{1, 4, 5, 7},
		{1, 4, 5, 6},
		{2, 7},
		{3},
		{1, 8},
		{2, 5, 9},
		{2, 5},
		{5, 6, 8},
	}
	if got := candidates(rgn5); !reflect.DeepEqual(wantRgn, got) {
		t.Errorf("want (region): %v, got: %v", wantRgn, got)
	}
}

func TestNakedPairs(t *testing.T) {
	g := NewGrid()

	// the candidates 6 & 8 in columns six and seven form a Naked Pair within
	// the row. Therefore, since one of these cells must be the 6 and the other
	// must be the 8, candidates 6 & 8 can be excluded from all other cells in
	// the row.
	row5 := g.row(5)
	setCandidates(row5, [GridSize][]byte{
		{7},
		{1, 2, 4, 5},
		{1, 4, 5},
		{2, 4, 5},
		{9},
		{6, 8},
		{6, 8},
		{1, 6, 8}, // 6 and 8 cannot be here
		{3},
	})

	nakedPairs(row5)

	want := [GridSize][]byte{
		{7},
		{1, 2, 4, 5},
		{1, 4, 5},
		{2, 4, 5},
		{9},
		{6, 8},
		{6, 8},
		{1},
		{3},
	}
	if got := candidates(row5); !reflect.DeepEqual(want, got) {
		t.Errorf("want (row): %v, got: %v", want, got)
	}
}

func TestHiddenPairs(t *testing.T) {
	g := NewGrid()

	// the candidates 1 & 9 are only located in two commented cells of a region,
	// and therefore form a 'hidden' pair. All candidates except 1 & 9 can
	// safely be excluded from these two cells as one cell must be the 1 while
	// the other must be the 9.
	rgn0 := g.region(0)
	setCandidates(rgn0, [GridSize][]byte{
		{7},
		{2, 3},
		{2, 3, 5},
		{1, 2, 5, 9}, // hidden pair: 1 & 9
		{8},
		{4},
		{1, 5, 9}, // hidden pair: 1 & 9
		{6},
		{3, 5},
	})

	hiddenPairs(rgn0)

	want := [GridSize][]byte{
		{7},
		{2, 3},
		{2, 3, 5},
		{1, 9},
		{8},
		{4},
		{1, 9},
		{6},
		{3, 5},
	}
	if got := candidates(rgn0); !reflect.DeepEqual(want, got) {
		t.Errorf("want (region): %v, got: %v", want, got)
	}
}
