package main

import (
	"bufio"
	"log"
	"os"
	"testing"
)

func BenchmarkSolver1(b *testing.B) {
	s, _ := NewSudoku("....7..2.8.......6.1.2.5...9.54....8.........3....85.1...3.2.8.4.......9.7..6....")

	for i := 0; i < b.N; i++ {
		s.Solve()
	}
}
func BenchmarkSolver2(b *testing.B) {
	s, _ := NewSudoku("..5...987.4..5...1..7......2...48....9.1.....6..2.....3..6..2.......9.7.......5..")
	for i := 0; i < b.N; i++ {
		s.Solve()
	}
}
func BenchmarkSolver3(b *testing.B) {
	s, _ := NewSudoku("38.6.......9.......2..3.51......5....3..1..6....4......17.5..8.......9.......7.32")
	for i := 0; i < b.N; i++ {
		s.Solve()
	}
}

func BenchmarkValidator1(b *testing.B) {
	s, _ := NewSudoku("....7..2.8.......6.1.2.5...9.54....8.........3....85.1...3.2.8.4.......9.7..6....")
	for i := 0; i < b.N; i++ {
		s.Validate()
	}
}

func BenchmarkValidator2(b *testing.B) {
	s, _ := NewSudoku("...92......68.3...19..7...623..4.1....1...7....8.3..297...8..91...5.72......64...")
	for i := 0; i < b.N; i++ {
		s.Validate()
	}
}

func BenchmarkValidator3(b *testing.B) {
	s, _ := NewSudoku("7..1523........92....3.....1....47.8.......6............9...5.6.4.9.7...8....6.1.")
	for i := 0; i < b.N; i++ {
		s.Validate()
	}
}

func TestValidSudoku(t *testing.T) {
	_, err := NewSudoku("7..1523........92....3.....1....47.8.......6............9...5.6.4.9.7...8....6.1.")
	if err != nil {
		t.Errorf("Invalid input %v", err)
	}

}

func TestInvalidSudoku(t *testing.T) {
	_, err := NewSudoku("7X.1523........92....3.....1....47.8.......6............9...5.6.4.9.7...8....6.1.")
	if err == nil {
		t.Errorf("Parsing invalid input")
	}

}

func TestSudoku(t *testing.T) {
	testFiles := []string{"hardest.txt", "top95.txt"}
	var str string
	for _, path := range testFiles {
		file, err := os.Open("input/" + path)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		line := 0
		for scanner.Scan() {
			line++
			str = scanner.Text()
			s, _ := NewSudoku(str)
			//s.Print()
			s.Solve()
			//s.Print()
			if !s.Validate() {
				t.Errorf("Invalid solution for %s:%d\n%s\n", path, line, str)
			}

		}

		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}

	}
}
