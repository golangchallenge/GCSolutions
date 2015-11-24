package main

import (
	"reflect"
	"testing"
)

func TestSolve(t *testing.T) {
	inputs := []string{
		// (a ∨ ¬b) ∧ (a ∨ b) should return a: True, b: anything
		`p cnf 2 2
		 1 -2 0
		 1 2 0`,
		// (a ∧ b) ∧ (a ∧ ¬b) should return: null
		`p cnf 2 4
		 1 0
		 2 0
		 1 0
		 -2 0`,
		// (a ∧ b) ∧ (¬b ∨ c) should return: a: True, b: True, c: True.
		`p cnf 3 3
		 1 0
		 2 0
		 -1 3 0`,
		// (x ∨ x ∨ y) ∧ (¬x ∨ ¬y ∨ ¬y) ∧ (¬x ∨ y ∨ y) x: False, y: True
		`p cnf 2 3
		 1 1 2 0
		 -1 -2 -2 0
		 -1 2 2 0`,
	}

	expecteds := [][][]SetVar{
		// (a ∨ ¬b) ∧ (a ∨ b) should return a: True, b: anything
		{
			{SetVar{VarNum: 1, Value: true}, SetVar{VarNum: 2, Value: false}},
			{SetVar{VarNum: 1, Value: true}, SetVar{VarNum: 2, Value: true}},
		},
		// (a ∧ b) ∧ (a ∧ ¬b) should return: null
		nil,
		// (a ∧ b) ∧ (¬b ∨ c) should return: a: True, b: True, c: True.
		{
			{SetVar{VarNum: 3, Value: true}, SetVar{VarNum: 2, Value: true}, SetVar{VarNum: 1, Value: true}},
		},
		// (x ∨ x ∨ y) ∧ (¬x ∨ ¬y ∨ ¬y) ∧ (¬x ∨ y ∨ y) x: False, y: True
		{
			{SetVar{VarNum: 1, Value: false}, SetVar{VarNum: 2, Value: true}},
		},
	}

	for i, input := range inputs {
		if i < len(inputs)-1 {
			continue
		}
		expectedSlns := expecteds[i]

		slns := testInput(t, input)

		if slns == nil || len(slns) == 0 {
			if expectedSlns != nil {
				t.Fatalf("test idx: %d. no solution expected, actual: %#v", i, slns)
			}
			continue
		}

		if len(slns) != len(expectedSlns) {
			t.Fatalf("test idx: %d. expected %d solutions, actual: %d - %#v", i, len(expectedSlns), len(slns), slns)
		}

		for j, expected := range expectedSlns {
			actual := slns[j].SetVars

			if len(actual) != len(expected) {
				t.Fatalf("test idx: %d. sln: %d/%d. expected: %#v actual: %#v", i, j+1, len(expectedSlns), expected, actual)
			}

			for _, e := range expected {
				found := false
				for _, a := range actual {
					if a.Value == e.Value {
						found = true
						if a.VarNum != a.VarNum {
							t.Fatalf("test idx: %d. sln: %d/%d. expected: %d %t actual: %d %t", i, j+1, len(expectedSlns), e.VarNum, e.Value, a.VarNum, a.Value)
						}
					}
				}
				if !found {
					t.Fatalf("test idx: %d. sln: %d/%d. expected: %d %t actual: not found", i, j+1, len(expectedSlns), e.VarNum, e.Value)
				}
			}
		}
	}
}

func testInput(t *testing.T, input string) []*SAT {
	sat, err := NewSAT(input, true, 100)
	if err != nil {
		t.Fatal(err)
	}
	sln := sat.Solve()
	return sln
}

func TestHasClause(t *testing.T) {
	clauseIntArray := []int{0, 2, 6, 8, 12}
	clause := intArrayToBin(clauseIntArray)
	for idx, val := range clauseIntArray {
		actual := indexOfValue(&clause, uint64(val))
		if actual != idx {
			t.Fatalf("%d not found in clause %b %b. idx:%d", val, clause[0], clause[1], idx)
		}
	}

	clauseIntArray = []int{0, 2, 6, 8, 12, 14}
	clause = intArrayToBin(clauseIntArray)
	for idx, val := range clauseIntArray {
		actual := indexOfValue(&clause, uint64(val))
		if actual != idx {
			t.Fatalf("%d not found in clause %b %b. idx:%d", val, clause[0], clause[1], idx)
		}
	}

	clauseIntArray = []int{-937, -737}
	clause = intArrayToBin(clauseIntArray)
	for idx, val := range clauseIntArray {
		uintVal := uint64(abs(val))
		if val < 0 {
			uintVal |= 0x400
		}

		actual := indexOfValue(&clause, uintVal)
		if actual != idx {
			t.Fatalf("%d (%b) not found in clause %b %b. idx:%d", val, uintVal, clause[0], clause[1], idx)
		}
	}
}

func TestUnitPropogation(t *testing.T) {
	// clause: 1 2 6 8 12
	clause := intArrayToBin([]int{1, 2, 6, 8, 12})

	expected := intArrayToBin([]int{1, 2, 6, 8})
	actual := up(&clause, 12, false)
	if !reflect.DeepEqual(&expected, actual) {
		t.Fatalf("expected: %v actual: %v", expected, actual)
	}

	expected = *satisfied
	actual = up(&clause, 12, true)
	if !reflect.DeepEqual(&expected, actual) {
		t.Fatalf("expected: %v actual: %v", expected, actual)
	}

	// clause: -1 2 6 8 12
	clause = intArrayToBin([]int{-1, 2, 6, 8, 12})

	expected = intArrayToBin([]int{2, 6, 8, 12})
	actual = up(&clause, 1, true)
	if !reflect.DeepEqual(&expected, actual) {
		t.Fatalf("expected: %v actual: %v", expected, actual)
	}

	expected = *satisfied
	actual = up(&clause, 1, false)
	if !reflect.DeepEqual(&expected, actual) {
		t.Fatalf("expected: %v actual: %v", expected, actual)
	}
}

func TestIntArrayToBin(t *testing.T) {
	// TODO
}
