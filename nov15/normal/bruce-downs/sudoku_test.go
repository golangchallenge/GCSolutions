package main

import (
	"bufio"
	"bytes"
	"flag"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"
)

func solveOfficial(tb testing.TB) {
	test :=
		`1 _ 3 _ _ 6 _ 8 _
_ 5 _ _ 8 _ 1 2 _
7 _ 9 1 _ 3 _ 5 6
_ 3 _ _ 6 7 _ 9 _
5 _ 7 8 _ _ _ 3 _
8 _ 1 _ 3 _ 5 _ 7
_ 4 _ _ 7 8 _ 1 _
6 _ 8 _ _ 2 _ 4 _
_ 1 2 _ 4 5 _ 7 8
`
	expect :=
		`1 2 3 4 5 6 7 8 9
4 5 6 7 8 9 1 2 3
7 8 9 1 2 3 4 5 6
2 3 4 5 6 7 8 9 1
5 6 7 8 9 1 2 3 4
8 9 1 2 3 4 5 6 7
3 4 5 6 7 8 9 1 2
6 7 8 9 1 2 3 4 5
9 1 2 3 4 5 6 7 8
`

	var puzz puzzle
	in := bytes.NewBufferString(test)

	if err := puzz.init(in); err != nil {
		tb.Error(err)
	}

	vout := &bytes.Buffer{}

	var category string
	puzz, category = solve(puzz, vout)
	tb.Log(vout)

	if err := puzz.solved(); err != nil {
		tb.Errorf("puzzle not solved [%s]", err)
	}
	if !strings.EqualFold("hard", category) {
		tb.Error("puzzle not hard")
	}

	out := &bytes.Buffer{}
	puzz.printme(out)

	actual := out.String()
	if expect != actual {
		tb.Errorf("Expect [%s] Actual [%s]", expect, actual)
	}

	tb.Log(out)
}

func TestOfficial(t *testing.T) {
	solveOfficial(t)
}

func BenchmarkOfficial(b *testing.B) {
	solveOfficial(b)
}

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

func TestDryRun(t *testing.T) {
	test :=
		`1 2 3 4 5 6 7 8 9
1 2 3 4 5 6 7 8 9
1 2 3 4 5 6 7 8 9
1 2 3 4 5 6 7 8 9
1 2 3 4 5 6 7 8 9
1 2 3 4 5 6 7 8 9
1 2 3 4 5 6 7 8 9
1 2 3 4 5 6 7 8 9
1 2 3 4 5 6 7 8 9
`

	var puzz puzzle
	in := bytes.NewBufferString(test)

	if err := puzz.init(in); err != nil {
		t.Error(err)
	}

	puzz.printme(ioutil.Discard)
}

func TestPuzzleEasy(t *testing.T) {
	var tests = []struct {
		test   string
		expect string
	}{
		{
			test: `_ 6 7 1 _ _ _ _ _
3 _ _ 5 _ 6 4 _ _
_ 1 _ 2 _ _ _ 9 6
_ 3 _ 7 4 _ _ _ 8
_ 7 2 _ _ _ 6 5 _
8 _ _ _ 6 2 _ 7 4
9 2 _ _ _ 7 _ 4 _
_ _ 8 4 _ 9 _ _ 7
_ _ _ _ _ 3 5 6 _
`,
			expect: `2 6 7 1 9 4 3 8 5
3 8 9 5 7 6 4 1 2
5 1 4 2 3 8 7 9 6
1 3 6 7 4 5 9 2 8
4 7 2 9 8 1 6 5 3
8 9 5 3 6 2 1 7 4
9 2 3 6 5 7 8 4 1
6 5 8 4 1 9 2 3 7
7 4 1 8 2 3 5 6 9
`,
		},
		{
			test: `_ _ 6 4 5 _ 1 8 _
_ _ _ _ _ 1 3 _ _
_ _ 1 _ 3 6 _ 2 _
1 9 _ 7 _ _ 2 _ 4
_ 3 _ _ 4 _ _ 6 _
7 _ 4 _ _ 8 _ 3 1
_ 5 _ 8 1 _ 4 _ _
_ _ 2 3 _ _ _ _ _
_ 4 3 _ 7 5 8 _ _
`,
			expect: `3 2 6 4 5 7 1 8 9
5 7 9 2 8 1 3 4 6
4 8 1 9 3 6 5 2 7
1 9 8 7 6 3 2 5 4
2 3 5 1 4 9 7 6 8
7 6 4 5 2 8 9 3 1
6 5 7 8 1 2 4 9 3
8 1 2 3 9 4 6 7 5
9 4 3 6 7 5 8 1 2
`,
		},
	}

	for _, test := range tests {
		var puzz puzzle
		in := bytes.NewBufferString(test.test)

		if err := puzz.init(in); err != nil {
			t.Error(err)
		}

		var category string
		vout := &bytes.Buffer{}
		puzz, category = solve(puzz, vout)
		t.Log(vout)

		if err := puzz.solved(); err != nil {
			t.Errorf("puzzle not solved [%s]", err)
		}
		if !strings.EqualFold(category, "easy") {
			t.Error("puzzle not easy")
		}

		out := &bytes.Buffer{}
		puzz.printme(out)

		actual := out.String()
		if test.expect != actual {
			t.Errorf("Expect [%s] Actual [%s]", test.expect, actual)
		}

		t.Log(out)
	}
}

func TestPuzzleMedium(t *testing.T) {
	var tests = []struct {
		test   string
		expect string
	}{
		{
			test: `_ 1 8 _ _ _ _ _ 2
_ _ _ 6 1 _ 9 _ _
5 _ _ _ _ 9 7 1 _
6 _ 5 _ 3 _ _ _ 9
_ 9 _ 2 _ 7 _ 4 _
2 _ _ _ 4 _ 3 _ 5
_ 6 2 4 _ _ _ _ 7
_ _ 4 _ 9 2 _ _ _
1 _ _ _ _ _ 4 2 _
`,
			expect: `9 1 8 5 7 4 6 3 2
4 2 7 6 1 3 9 5 8
5 3 6 8 2 9 7 1 4
6 4 5 1 3 8 2 7 9
8 9 3 2 5 7 1 4 6
2 7 1 9 4 6 3 8 5
3 6 2 4 8 1 5 9 7
7 5 4 3 9 2 8 6 1
1 8 9 7 6 5 4 2 3
`,
		},
	}

	for _, test := range tests {
		var puzz puzzle
		in := bytes.NewBufferString(test.test)

		if err := puzz.init(in); err != nil {
			t.Error(err)
		}

		var category string
		vout := &bytes.Buffer{}
		puzz, category = solve(puzz, vout)
		t.Log(vout)

		if err := puzz.solved(); err != nil {
			t.Errorf("puzzle not solved [%s]", err)
		}
		if !strings.EqualFold(category, "medium") {
			t.Error("puzzle not medium")
		}

		out := &bytes.Buffer{}
		puzz.printme(out)

		actual := out.String()
		if test.expect != actual {
			t.Errorf("Expect [%s] Actual [%s]", test.expect, actual)
		}

		t.Log(out)
	}
}

func TestPuzzleHard(t *testing.T) {
	var tests = []struct {
		test   string
		expect string
	}{
		{
			test: `_ _ _ _ _ _ 7 9 _
4 2 _ 3 _ _ _ _ 1
_ 6 5 _ 7 _ _ _ _
_ 3 9 4 _ _ _ 7 _
1 4 _ _ _ _ _ 2 8
_ 7 _ _ _ 1 9 4 _
_ _ _ _ 9 _ 8 5 _
8 _ _ _ _ 5 _ 6 2
_ 5 2 _ _ _ _ _ _
`,
			expect: `3 8 1 6 2 4 7 9 5
4 2 7 3 5 9 6 8 1
9 6 5 1 7 8 2 3 4
5 3 9 4 8 2 1 7 6
1 4 6 9 3 7 5 2 8
2 7 8 5 6 1 9 4 3
6 1 4 2 9 3 8 5 7
8 9 3 7 1 5 4 6 2
7 5 2 8 4 6 3 1 9
`,
		},
	}

	for _, test := range tests {
		var puzz puzzle
		in := bytes.NewBufferString(test.test)

		if err := puzz.init(in); err != nil {
			t.Error(err)
		}

		var category string
		vout := &bytes.Buffer{}
		puzz, category = solve(puzz, vout)
		t.Log(vout)

		if err := puzz.solved(); err != nil {
			t.Errorf("puzzle not solved [%s]", err)
		}
		if !strings.EqualFold(category, "hard") {
			t.Error("puzzle not hard")
		}

		out := &bytes.Buffer{}
		puzz.printme(out)

		actual := out.String()
		if test.expect != actual {
			t.Errorf("Expect [%s] Actual [%s]", test.expect, actual)
		}

		t.Log(out)
	}
}

func TestIOInvalidLines(t *testing.T) {
	tests := []string{
		// 8 lines
		`1 _ 3 _ _ 6 _ 8 _
_ 5 _ _ 8 _ 1 2 _
7 _ 9 1 _ 3 _ 5 6
_ 3 _ _ 6 7 _ 9 _
5 _ 7 8 _ _ _ 3 _
8 _ 1 _ 3 _ 5 _ 7
_ 4 _ _ 7 8 _ 1 _
6 _ 8 _ _ 2 _ 4 _
`,
		// 10 lines
		`1 _ 3 _ _ 6 _ 8 _
_ 5 _ _ 8 _ 1 2 _
7 _ 9 1 _ 3 _ 5 6
_ 3 _ _ 6 7 _ 9 _
5 _ 7 8 _ _ _ 3 _
8 _ 1 _ 3 _ 5 _ 7
_ 4 _ _ 7 8 _ 1 _
6 _ 8 _ _ 2 _ 4 _
_ 1 2 _ 4 5 _ 7 8
1 2 3 4 5 6 7 8 9
`,
		// missing last line feed
		`1 _ 3 _ _ 6 _ 8 _
_ 5 _ _ 8 _ 1 2 _
7 _ 9 1 _ 3 _ 5 6
_ 3 _ _ 6 7 _ 9 _
5 _ 7 8 _ _ _ 3 _
8 _ 1 _ 3 _ 5 _ 7
_ 4 _ _ 7 8 _ 1 _
6 _ 8 _ _ 2 _ 4 _
_ 1 2 _ 4 5 _ 7 8`,
		// empty line
		``}

	for _, test := range tests {
		var puzz puzzle
		in := bytes.NewBufferString(test)

		if err := puzz.init(in); err == nil {
			t.Error("error expected initializing puzzle")
		}
	}
}

func TestIOInvalidFieldCount(t *testing.T) {
	tests := []string{
		// 8 fields in first line
		`1 _ 3 _ _ 6 _ 8
_ 5 _ _ 8 _ 1 2 _
7 _ 9 1 _ 3 _ 5 6
_ 3 _ _ 6 7 _ 9 _
5 _ 7 8 _ _ _ 3 _
8 _ 1 _ 3 _ 5 _ 7
_ 4 _ _ 7 8 _ 1 _
6 _ 8 _ _ 2 _ 4 _
_ 1 2 _ 4 5 _ 7 8
`,
		// 8 fields in last line
		`1 _ 3 _ _ 6 _ 8 _
_ 5 _ _ 8 _ 1 2 _
7 _ 9 1 _ 3 _ 5 6
_ 3 _ _ 6 7 _ 9 _
5 _ 7 8 _ _ _ 3 _
8 _ 1 _ 3 _ 5 _ 7
_ 4 _ _ 7 8 _ 1 _
6 _ 8 _ _ 2 _ 4 _
_ 1 2 _ 4 5 _ 7
`,
		// 10 fields in first line
		`1 _ 3 _ _ 6 _ 8 _ 1
_ 5 _ _ 8 _ 1 2 _
7 _ 9 1 _ 3 _ 5 6
_ 3 _ _ 6 7 _ 9 _
5 _ 7 8 _ _ _ 3 _
8 _ 1 _ 3 _ 5 _ 7
_ 4 _ _ 7 8 _ 1 _
6 _ 8 _ _ 2 _ 4 _
_ 1 2 _ 4 5 _ 7 8
`,
		// 10 fields in last line
		`1 _ 3 _ _ 6 _ 8 _
_ 5 _ _ 8 _ 1 2 _
7 _ 9 1 _ 3 _ 5 6
_ 3 _ _ 6 7 _ 9 _
5 _ 7 8 _ _ _ 3 _
8 _ 1 _ 3 _ 5 _ 7
_ 4 _ _ 7 8 _ 1 _
6 _ 8 _ _ 2 _ 4 _
_ 1 2 _ 4 5 _ 7 8 1
`}

	for _, test := range tests {
		var puzz puzzle
		in := bytes.NewBufferString(test)

		if err := puzz.init(in); err == nil {
			t.Error("error expected initializing puzzle")
		}
	}
}

func TestIOInvalidFieldRange(t *testing.T) {
	tests := []string{
		// field < 0
		`1 _ 3 _ _ 6 _ 8 -1
_ 5 _ _ 8 _ 1 2 _
7 _ 9 1 _ 3 _ 5 6
_ 3 _ _ 6 7 _ 9 _
5 _ 7 8 _ _ _ 3 _
8 _ 1 _ 3 _ 5 _ 7
_ 4 _ _ 7 8 _ 1 _
6 _ 8 _ _ 2 _ 4 _
_ 1 2 _ 4 5 _ 7 8
`,
		// field > 9
		`1 _ 3 _ _ 6 _ 8 10
_ 5 _ _ 8 _ 1 2 _
7 _ 9 1 _ 3 _ 5 6
_ 3 _ _ 6 7 _ 9 _
5 _ 7 8 _ _ _ 3 _
8 _ 1 _ 3 _ 5 _ 7
_ 4 _ _ 7 8 _ 1 _
6 _ 8 _ _ 2 _ 4 _
_ 1 2 _ 4 5 _ 7 8
`,
		// field = 12345
		`1 _ 3 _ _ 6 _ 8 _
_ 5 _ _ 8 _ 1 2 _
7 _ 9 1 _ 3 _ 5 6
_ 3 _ _ 6 7 _ 9 _
5 _ 7 8 _ _ _ 3 _
8 _ 1 _ 3 _ 5 _ 7
_ 4 _ _ 7 8 _ 1 _
6 _ 8 _ _ 2 _ 4 _
_ 1 2 _ 4 5 _ 7 12345
`,
		// field is alpha
		`1 _ 3 _ _ 6 _ 8 foobar
_ 5 _ _ 8 _ 1 2 _
7 _ 9 1 _ 3 _ 5 6
_ 3 _ _ 6 7 _ 9 _
5 _ 7 8 _ _ _ 3 _
8 _ 1 _ 3 _ 5 _ 7
_ 4 _ _ 7 8 _ 1 _
6 _ 8 _ _ 2 _ 4 _
_ 1 2 _ 4 5 _ 7 8
`}

	for _, test := range tests {
		var puzz puzzle
		in := bytes.NewBufferString(test)

		if err := puzz.init(in); err == nil {
			t.Error("error expected initializing puzzle")
		}
	}
}

func testAllPuzzles(t testing.TB) {
	filename := os.Getenv("SUDOKU_PUZZLE_FILENAME")
	if len(filename) < 1 {
		t.Skip("Provide SUDOKU_PUZZLE_FILENAME env to test all puzzles")
	}

	file, err := os.Open(filename)
	if err != nil {
		t.Error(err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	var total, longest time.Duration
	var count int

OUTER_LOOP:
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Error(err)
			break
		}
		if len(line) != 81+1 {
			t.Log("line length not 81")
			continue
		}

		in := &bytes.Buffer{}
		for k, v := range line {
			if k > 0 {
				if k%9 == 0 {
					in.WriteRune('\n')
				} else {
					in.WriteRune(' ')
				}
			}

			switch v {
			case '0', '.', '_':
				in.WriteRune('_')
			case '1', '2', '3', '4', '5', '6', '7', '8', '9':
				in.WriteRune(v)
			case '\n':
				// no-op
			default:
				t.Logf("Invalid field value [%s]", string(v))
				continue OUTER_LOOP
			}
		}

		var puzz puzzle
		if err := puzz.init(in); err != nil {
			t.Error(err)
			continue
		}

		before := time.Now()
		puzz, _ = solve(puzz, ioutil.Discard)
		duration := time.Now().Sub(before)

		total += duration
		if duration > longest {
			longest = duration
		}
		count++

		if err := puzz.solved(); err == nil {
			t.Logf("puzzle solved in %.3f [%s]", duration.Seconds(), line[:81])
		} else {
			t.Errorf("puzzle not solved in %.3f [%s] [%s]", duration.Seconds(), line[:81], err)
		}
	}

	t.Logf("%d puzzles solved in %.3f seconds [%.3f per] [%.3f longest]",
		count, total.Seconds(), total.Seconds()/float64(count), longest.Seconds())
}

func TestAllPuzzles(t *testing.T) {
	testAllPuzzles(t)
}

func BenchmarkAllPuzzles(b *testing.B) {
	testAllPuzzles(b)
}
