package parser

import (
	"strings"
	"testing"
)

func TestParseLineHappy(t *testing.T) {
	inputLine := "1 _ 3 _ _ 6 _ 8 _"
	outArray, err := parseLine(inputLine)
	if err != nil {
		t.Error("Unable to parse the input line. Parseline returned error", err)
		t.FailNow()
	}
	expectedArray := [9]int{1, 0, 3, 0, 0, 6, 0, 8, 0}
	for ind, item := range outArray {
		if item != expectedArray[ind] {
			t.Log("Expected value did not match the actual one. Actual and expected one are", outArray[ind], expectedArray[ind])
		}
	}
}

func TestParseLineBadInput(t *testing.T) {
	inputLine := "1 _ 3 _ _ 7 _ 8 &" + "\r\n"
	_, err := parseLine(inputLine)
	if err == nil {
		t.Error("Expected error, but did not get error. Failing this test case")
		t.FailNow()
	}
}

func TestGetInput(t *testing.T) {
	rawInputString := "1 _ 3 _ _ 6 _ 8 _" + "\n" +
		"_ 5 _ _ 8 _ 1 2 _" + "\n" +
		"7 _ 9 1 _ 3 _ 5 6" + "\n" +
		"_ 3 _ _ 6 7 _ 9 _" + "\n" +
		"5 _ 7 8 _ _ _ 3 _" + "\n" +
		"8 _ 1 _ 3 _ 5 _ 7" + "\n" +
		"_ 4 _ _ 7 8 _ 1 _" + "\n" +
		"6 _ 8 _ _ 2 _ 4 _" + "\n" +
		"_ 1 2 _ 4 5 _ 7 8" + "\n"

	stringReader := strings.NewReader(rawInputString)
	outGrid, err := GetInput(stringReader)
	if err != nil {
		t.Error("Unable to parse input grid. Error is ", err)
	}
	expectedArray1stRow := [9]int{1, 0, 3, 0, 0, 6, 0, 8, 0}
	if expectedArray1stRow != outGrid[0] {
		t.Error("Unexpected value for the first row of the array")
		t.Error("Expecter is: ", expectedArray1stRow)
		t.Error("Actual is ", outGrid[0])
	}
}
