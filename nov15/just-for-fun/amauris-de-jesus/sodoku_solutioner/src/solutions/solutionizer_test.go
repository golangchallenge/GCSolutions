package solutions;

import (
	"testing"
	"sodoku"
	"strings"
)

var (

	TABLE string = `5 3 _ _ 7 _ _ _ _
		6 _ _ 1 9 5 _ _ _
		_ 9 8 _ _ _ _ 6 _
		8 _ _ _ 6 _ _ _ 3
		4 _ _ 8 _ 3 _ _ 1
		7 _ _ _ 2 _ _ _ 6
		_ 6 _ _ _ _ 2 8 _
		_ _ _ 4 1 9 _ _ 5
		_ _ _ _ 8 _ _ 7 9`

	TABLE_ANSWER string = `5 3 4 6 7 8 9 1 2 
6 7 2 1 9 5 3 4 8 
1 9 8 3 4 2 5 6 7 
8 5 9 7 6 1 4 2 3 
4 2 6 8 5 3 7 9 1 
7 1 3 9 2 4 8 5 6 
9 6 1 5 3 7 2 8 4 
2 8 7 4 1 9 6 3 5 
3 4 5 2 8 6 1 7 9`

)

//helper functions

func buildSolutionizer() *Solutionizer {

	return &Solutionizer{}
}

func compactString(input string) string {

	compact := strings.Replace(input, " ", "", -1)
	compact = strings.Replace(compact, "\n", "", -1)
	compact = strings.Replace(compact, "\r", "", -1)

	return compact
}

//end of helper functions

//main entry function that calls the right
//functions for gettnig sodoku answer
func TestGetSodokuSolution(t *testing.T) {

	solutionizer := buildSolutionizer()
	board := sodoku.GetPreDefinedBoard(TABLE, 9)
	
	computedAnswer := solutionizer.GetSodokuSolution(board)

	compactAnswer := compactString(computedAnswer)
	compactCorrectAnswer := compactString(TABLE_ANSWER)


	if compactAnswer!=compactCorrectAnswer {
		t.Error("Expected ", compactCorrectAnswer, " GOT ", compactAnswer)
	}
}