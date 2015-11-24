package main

import (
	"bytes"
	"flag"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
)

func runMain(t *testing.T, input io.Reader, args ...string) (int, *bytes.Buffer) {
	defer func(fs *flag.FlagSet) { flag.CommandLine = fs }(flag.CommandLine)
	defer func(args []string) { os.Args = args }(os.Args)
	flag.CommandLine = flag.NewFlagSet("soodohkoo", 0)
	os.Args = append([]string{"soodohkoo"}, args...)

	wg := sync.WaitGroup{}

	if input == nil {
		input = bytes.NewBuffer(nil)
	}
	defer func(f *os.File) { os.Stdin = f }(os.Stdin)
	stdinR, stdinW, err := os.Pipe()
	if err != nil {
		t.Fatalf("error creating pipe: %s", err)
	}
	defer stdinR.Close()
	defer stdinW.Close()
	os.Stdin = stdinR
	wg.Add(1)
	go func() { io.Copy(stdinW, input); stdinW.Close(); wg.Done() }()

	defer func(f *os.File) { os.Stdout = f }(os.Stdout)
	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		t.Fatalf("error creating pipe: %s", err)
	}
	defer stdoutW.Close()
	defer stdoutR.Close()
	os.Stdout = stdoutW
	stdoutBuf := bytes.NewBuffer(nil)
	wg.Add(1)
	go func() { io.Copy(stdoutBuf, stdoutR); wg.Done() }()

	// for simplicity sake, merge stderr into stdout
	defer func(f *os.File) { os.Stderr = f }(os.Stderr)
	os.Stderr = os.Stdout

	status := mainMain()
	stdinR.Close()  // unblock STDIN io.Copy()
	stdoutW.Close() // unblock STDOUT io.Copy()
	wg.Wait()

	return status, stdoutBuf
}

func TestMainSolve(t *testing.T) {
	input := strings.NewReader(`_ 8 _ _ 6 _ _ _ _
5 4 _ _ _ 7 _ 3 _
_ _ _ 1 _ _ 8 6 7
_ _ 9 _ 3 _ _ _ 6
_ _ 5 _ _ _ 3 _ _
3 _ _ _ 4 _ 2 _ _
7 5 4 _ _ 6 _ _ _
_ 2 _ 4 _ _ _ 7 9
_ _ _ _ 2 _ _ 8 _
`)
	status, output := runMain(t, input, "-mode=solve", "-stats")
	if status != 0 {
		t.Errorf("main returned %d, expected %d", status, 0)
	}

	expectedOutput := `1 8 7 3 6 9 4 5 2
5 4 6 2 8 7 9 3 1
9 3 2 1 5 4 8 6 7
2 1 9 5 3 8 7 4 6
4 6 5 9 7 2 3 1 8
3 7 8 6 4 1 2 9 5
7 5 4 8 9 6 1 2 3
8 2 3 4 1 5 6 7 9
6 9 1 7 2 3 5 8 4
`
	if !strings.HasPrefix(output.String(), expectedOutput) {
		t.Errorf("output did not match solved board")
	}

	//TODO check stats?

	b := NewBoard()
	_, err := b.ReadFrom(output)
	if err != nil {
		t.Errorf("error reading output board: %s", err)
	}
	if !b.Solved() {
		t.Errorf("output board is not solved")
	}
}

func TestMainSolve_invalid(t *testing.T) {
	input := strings.NewReader(`X 8 _ _ 6 _ _ _ _
5 4 _ _ _ 7 _ 3 _
_ _ _ 1 _ _ 8 6 7
_ _ 9 _ 3 _ _ _ 6
_ _ 5 _ _ _ 3 _ _
3 _ _ _ 4 _ 2 _ _
7 5 4 _ _ 6 _ _ _
_ 2 _ 4 _ _ _ 7 9
_ _ _ _ 2 _ _ 8 _
`)
	status, output := runMain(t, input, "-mode=solve", "-stats")
	if status != 1 {
		t.Errorf("main returned %d, expected %d", status, 1)
	}
	if output.String() != "invalid byte\n" {
		t.Errorf("output is %q, expected %q", output.String(), "invalid byte\n")
	}
}

func TestMainSolve_noSolution(t *testing.T) {
	input := strings.NewReader(`_ 8 _ _ 6 _ _ _ 1
5 4 _ _ _ 7 _ 3 _
_ _ _ 1 _ _ 8 6 7
_ _ 9 _ 3 _ _ _ 6
_ _ 5 _ _ _ 3 _ _
3 _ _ _ 4 _ 2 _ _
7 5 4 _ _ 6 _ _ _
_ 2 _ 4 _ _ _ 7 9
_ _ _ _ 2 _ _ 8 _
`)
	status, output := runMain(t, input, "-mode=solve", "-stats")
	if status != 1 {
		t.Errorf("main returned %d, expected %d", status, 1)
	}
	if output.String() != "invalid board: no solution\n" {
		t.Errorf("output is %q, expected %q", output.String(), "invalid board: no solution\n")
	}
}

func TestMainSolveStream(t *testing.T) {
	input := strings.NewReader(`_ 8 _ _ 6 _ _ _ _
5 4 _ _ _ 7 _ 3 _
_ _ _ 1 _ _ 8 6 7
_ _ 9 _ 3 _ _ _ 6
_ _ 5 _ _ _ 3 _ _
3 _ _ _ 4 _ 2 _ _
7 5 4 _ _ 6 _ _ _
_ 2 _ 4 _ _ _ 7 9
_ _ _ _ 2 _ _ 8 _
_ _ _ _ 2 _ _ 8 _
_ 2 _ 4 _ _ _ 7 9
7 5 4 _ _ 6 _ _ _
3 _ _ _ 4 _ 2 _ _
_ _ 5 _ _ _ 3 _ _
_ _ 9 _ 3 _ _ _ 6
_ _ _ 1 _ _ 8 6 7
5 4 _ _ _ 7 _ 3 _
_ 8 _ _ 6 _ _ _ _
`)
	status, output := runMain(t, input, "-mode=solveStream")
	if status != 0 {
		t.Errorf("main returned %d, expected %d", status, 0)
	}

	b := NewBoard()
	_, err := b.ReadFrom(output)
	if err != nil {
		t.Errorf("error reading output board: %s", err)
	}
	if !b.Solved() {
		t.Errorf("output board is not solved")
	}

	b = NewBoard()
	_, err = b.ReadFrom(output)
	if err != nil {
		t.Errorf("error reading output board: %s", err)
	}
	if !b.Solved() {
		t.Errorf("output board is not solved")
	}
}

func TestMainGenerate(t *testing.T) {
	status, output := runMain(t, nil, "-mode=generate", "-difficulty=easy")
	if status != 0 {
		t.Errorf("main returned %d, expected %d", status, 0)
	}

	b := NewBoard()
	_, err := b.ReadFrom(output)
	if err != nil {
		t.Errorf("error reading output board: %s", err)
	}

	unknownCount := 0
	for _, t := range b.Tiles {
		if !t.isKnown() {
			unknownCount++
		}
	}
	if unknownCount != difficulties["easy"] {
		t.Errorf("have %d unknown tiles, expected %d", unknownCount, difficulties["easy"])
	}
}

func TestMainGenerate_difficultyInt(t *testing.T) {
	status, output := runMain(t, nil, "-mode=generate", "-difficulty=3")
	if status != 0 {
		t.Errorf("main returned %d, expected %d", status, 0)
	}

	b := NewBoard()
	_, err := b.ReadFrom(output)
	if err != nil {
		t.Errorf("error reading output board: %s", err)
	}

	unknownCount := 0
	for _, t := range b.Tiles {
		if !t.isKnown() {
			unknownCount++
		}
	}
	if unknownCount != 3 {
		t.Errorf("have %d unknown tiles, expected %d", unknownCount, 3)
	}
}
