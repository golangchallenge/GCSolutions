package main

import (
	"reflect"
	"testing"
)

func TestUnion(t *testing.T) {
	// arrange
	inputs := [][][]int{
		{{1, 2, 3}, {4, 5, 6}},
		{{1, 2, 3}, {3, 2, 1}},
		{{1, 2, 3}, {2, 2, 4}},
		{{1, 2, 3}, {3, 4, 5}},
	}

	expecteds := [][]int{
		{1, 2, 3, 4, 5, 6},
		{1, 2, 3},
		{1, 2, 3, 4},
		{1, 2, 3, 4, 5},
	}

	// act
	var actuals [][]int
	for _, input := range inputs {
		actual := union(input[0], input[1])
		actuals = append(actuals, actual)
	}

	// assert
	for i, expected := range expecteds {
		actual := actuals[i]
		if !reflect.DeepEqual(expected, actual) {
			t.Fatalf("inputs: %v expected: %v actual: %v", inputs[i], expected, actual)
		}
	}
}

func TestIntersect(t *testing.T) {
	// arrange
	inputs := [][][]int{
		{{1, 2, 3}, {4, 5, 6}},
		{{1, 2, 3}, {3, 2, 1}},
		{{1, 2, 3}, {2, 2, 4}},
		{{1, 2, 3}, {3, 4, 5}},
	}

	expecteds := [][]int{
		{},
		{1, 2, 3},
		{2},
		{3},
	}

	// act
	var actuals [][]int
	for _, input := range inputs {
		actual := intersect(input[0], input[1])
		actuals = append(actuals, actual)
	}

	// assert
	for i, expected := range expecteds {
		actual := actuals[i]
		if len(expected) == 0 && len(actual) == 0 {
			continue
		}
		if !reflect.DeepEqual(expected, actual) {
			t.Fatalf("inputs: %v expected: %v actual: %v", inputs[i], expected, actual)
		}
	}
}

func TestSubtract(t *testing.T) {
	// arrange
	inputs := [][][]int{
		{{1, 2, 3}, {4, 5, 6}},
		{{1, 2, 3}, {3, 2, 1}},
		{{1, 2, 3}, {2, 2, 4}},
		{{1, 2, 3}, {3, 4, 5}},
	}

	expecteds := [][]int{
		{1, 2, 3},
		{},
		{1, 3},
		{1, 2},
	}

	// act
	var actuals [][]int
	for _, input := range inputs {
		actual := subtract(input[0], input[1])
		actuals = append(actuals, actual)
	}

	// assert
	for i, expected := range expecteds {
		actual := actuals[i]
		if len(expected) == 0 && len(actual) == 0 {
			continue
		}
		if !reflect.DeepEqual(expected, actual) {
			t.Fatalf("inputs: %v expected: %v actual: %v", inputs[i], expected, actual)
		}
	}
}
