package main

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"
)

func TestSudoku(t *testing.T) {

	const puzzleDir = "puzzles"

	puzzleFileInfos, err := ioutil.ReadDir(puzzleDir)
	if err != nil {
		t.Fatal(err)
	}

	for _, pfi := range puzzleFileInfos {
		pname := path.Join(puzzleDir, pfi.Name())
		f, err := os.Open(pname)
		if err != nil {
			t.Fatal(err)
		}
		b, err := BoardFromReader(f)
		if err != nil {
			t.Fatal(err)
		}
		tstart := time.Now()
		if solved := b.Solve(); solved {
			tend := time.Now()
			t.Logf("solved %v in %v\n%v", pname, tend.Sub(tstart), b)
		} else {
			t.Fatalf("failed to solve %v\n%v", pname, b)
		}
	}
}

func TestUnsolvableSudoku(t *testing.T) {

	const puzzleDir = "unsolvable_puzzles"

	puzzleFileInfos, err := ioutil.ReadDir(puzzleDir)
	if err != nil {
		t.Fatal(err)
	}

	for _, pfi := range puzzleFileInfos {
		pname := path.Join(puzzleDir, pfi.Name())
		f, err := os.Open(pname)
		if err != nil {
			t.Fatal(err)
		}
		b, err := BoardFromReader(f)
		if err != nil {
			t.Fatal(err)
		}
		tstart := time.Now()
		if solved := b.Solve(); !solved {
			tend := time.Now()
			t.Logf("failed to solve %v in %v\n%v", pname, tend.Sub(tstart), b)
		} else {
			t.Fatalf("found an impossible solution for %v\n%v", pname, b)
		}
	}
}

func TestInvalidInputs(t *testing.T) {

	const puzzleDir = "invalid_inputs"

	puzzleFileInfos, err := ioutil.ReadDir(puzzleDir)
	if err != nil {
		t.Fatal(err)
	}

	for _, pfi := range puzzleFileInfos {
		pname := path.Join(puzzleDir, pfi.Name())
		f, err := os.Open(pname)
		if err != nil {
			t.Fatal(err)
		}
		b, err := BoardFromReader(f)
		if err != nil {
			t.Logf("failed to read puzzle from %v, err: %v", pname, err)
		} else {
			t.Fatalf("read a puzzle from invalid file %v:\n%v", pname, b)
		}
	}
}

func TestSudokuMain(t *testing.T) {

	const (
		outfile = "main_stdout.txt"
		errfile = "main_stderr.txt"
	)
	puzzles := []string{
		"puzzles/gochallenge_example.txt",
		"invalid_inputs/junk.txt",
		"unsolvable_puzzles/broken.txt",
	}

	oldStdin := os.Stdin
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	newStdout, err := os.Create(outfile)
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = newStdout

	newStderr, err := os.Create(errfile)
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = newStderr

	for _, p := range puzzles {
		f, err := os.Open(p)
		if err != nil {
			t.Fatal(err)
		}
		os.Stdin = f
		mainExitCode()
	}

	os.Stdin = oldStdin
	os.Stdout = oldStdout
	os.Stderr = oldStderr
	os.Remove(outfile)
	os.Remove(errfile)
}

func BenchmarkSudoku(b *testing.B) {

	const puzzle = "puzzles/top95_21.txt"

	f, err := os.Open(puzzle)
	if err != nil {
		b.Fatal(err)
	}

	board, err := BoardFromReader(f)
	if err != nil {
		b.Fatal(err)
	}

	// solve a new copy of the board b.N times
	for i := 0; i < b.N; i++ {
		freshBoard := board
		freshBoard.Solve()
	}
}
