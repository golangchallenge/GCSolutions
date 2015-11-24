package main

import (
	"testing"
	"fmt"
)

func TestGetSegmentNumber(t *testing.T) {

	s := getSegmentNumber(0)
	if (0 != s) {
		t.Errorf("0 Should be segment 0, got %d", s)
	}
	s = getSegmentNumber(2)
	if (0 != s) {
		t.Errorf("2 Should be segment 0, got %d", s)
	}
	s = getSegmentNumber(3)
	if (1 != s) {
		t.Errorf("3 Should be segment 1, got %d", s)
	}
	s = getSegmentNumber(9)
	if (0 != s) {
		t.Errorf("9 Should be segment 0, got %d", s)
	}
	s = getSegmentNumber(41)
	if (4 != s) {
		t.Errorf("41 Should be segment 4, got %d", s)
	}
	s = getSegmentNumber(80)
	if (8 != s) {
		t.Errorf("80 Should be segment 8, got %d", s)
	}
}

func TestGetRow(t *testing.T) {
	expected := [9]int8{1, 0, 3, 0, 0, 6, 0, 8, 0}
	result := getRow(examplePuzzle, 1)
	if expected != result {
		t.Errorf("Failed\nexp: %s\ngot: %s", rowToString(expected), rowToString(result))
	}

	expected = [9]int8{7, 0, 9, 1, 0, 3, 0, 5, 6}
	result = getRow(examplePuzzle, 23)
	if expected != result {
		t.Errorf("Failed\nexp: %s\ngot: %s", rowToString(expected), rowToString(result))
	}
}

func TestGetCol(t *testing.T) {
	expected := [9]int8{
		1,
		0,
		7,
		0,
		5,
		8,
		0,
		6,
		0,
	}
	result := getCol(examplePuzzle, 18)
	if expected != result {
		t.Errorf("Failed\nexp: %s\ngot: %s", rowToString(expected), rowToString(result))
	}

	expected = [9]int8{
		8,
		2,
		5,
		9,
		3,
		0,
		1,
		4,
		7,
	}
	result = getCol(examplePuzzle, 25)
	if expected != result {
		t.Errorf("Failed\nexp: %s\ngot: %s", rowToString(expected), rowToString(result))
	}
}

func TestGetBlock(t *testing.T) {
	expected := [9]int8{
		1, 0, 3,
		0, 5, 0,
		7, 0, 9,
	}
	result := getBlock(examplePuzzle, 10)
	if expected != result {
		t.Errorf("Failed\nexp: %s\ngot: %s", rowToString(expected), rowToString(result))
	}
	expected = [9]int8{
		0, 9, 0,
		0, 3, 0,
		5, 0, 7,
	}
	result = getBlock(examplePuzzle, 35)
	if expected != result {
		t.Errorf("Failed\nexp: %s\ngot: %s", rowToString(expected), rowToString(result))
	}
	expected = [9]int8{
		0, 0, 6,
		0, 8, 0,
		1, 0, 3,
	}
	result = getBlock(examplePuzzle, 4)
	if expected != result {
		t.Errorf("Failed\nexp: %s\ngot: %s", rowToString(expected), rowToString(result))
	}
}


func rowToString(r [9]int8) string {
	var str string = ""
	for i := 0; i < 9; i++ {
		str = fmt.Sprintf("%s %d", str, r[i])
	}
	return str
}


func TestCountUnused(t *testing.T) {
	expected := 1
	result := countSetBits(2)
	if expected != result {
		t.Errorf("Failed exp: %d got: %d", expected, result)
	}

	expected = 2
	result = countSetBits(3)
	if expected != result {
		t.Errorf("Failed exp: %d got: %d", expected, result)
	}

	expected = 2
	result = countSetBits(36)
	if expected != result {
		t.Errorf("Failed exp: %d got: %d", expected, result)
	}
}

func TestAvailableSet(t *testing.T) {
	// 16 ... 10  9  8  7  6  5  4  3  2  1
	//  0 ...  0  1  1  1  1  1  1  1  1  1 = 1,2,3,4,5,6,7,8,9 = 511
	//  0 ...  0  0  0  0  0  1  0  0  1  0 = 2,5 (set1)
	//  0 ...  0  0  0  0  1  0  1  0  0  1 = 1,4,6 (set2)
	//  0 ...  0  1  1  0  0  0  0  0  0  0 = 8,9 (set3)
	//  0 ...  0  0  0  1  0  0  0  1  0  0 = 3,7 (Available)

	expected := int16(68)
	result := availableSet(18,41,384)
	if expected != result {
		t.Errorf("Failed exp: %d got: %d", expected, result)
	}
}

func TestSetToNumber(t *testing.T) {
	//  0 ...  0  0  0  0  0  1  0  0  0  0 = 5 (set1)
	expected := int8(5)
	result := setToNumber(16)
	if expected != result {
		t.Errorf("Failed exp: %d got: %d", expected, result)
	}

	expected = int8(1)
	result = setToNumber(1)
	if expected != result {
		t.Errorf("Failed exp: %d got: %d", expected, result)
	}
	expected = int8(9)
	result = setToNumber(256)
	if expected != result {
		t.Errorf("Failed exp: %d got: %d", expected, result)
	}
}

func TestNumberInSet(t *testing.T) {
	expected := true
	result := numberInSet(9, 256)
	if expected != result {
		t.Errorf("Failed exp: %d got: %d", expected, result)
	}

	expected = false
	result = numberInSet(9, 255)
	if expected != result {
		t.Errorf("Failed exp: %d got: %d", expected, result)
	}
}

func Test5thNumber(t *testing.T) {
	puz := examplePuzzle;
	puz[1] = 2
	puz[3] = 4
	i := 4;
	available := availableSet(
		createSet(getBlock(puz, i)),
		createSet(getRow(puz, i)),
		createSet(getCol(puz, i)),
	)
	expected := int16(272);
	if available != expected {
		t.Errorf("Failed exp: %d got: %d", expected, available)
	}
}


var examplePuzzle = [81]int8{
	1, 0, 3, 0, 0, 6, 0, 8, 0,
	0, 5, 0, 0, 8, 0, 1, 2, 0,
	7, 0, 9, 1, 0, 3, 0, 5, 6,
	0, 3, 0, 0, 6, 7, 0, 9, 0,
	5, 0, 7, 8, 0, 0, 0, 3, 0,
	8, 0, 1, 0, 3, 0, 5, 0, 7,
	0, 4, 0, 0, 7, 8, 0, 1, 0,
	6, 0, 8, 0, 0, 2, 0, 4, 0,
	0, 1, 2, 0, 4, 5, 0, 7, 8,
}