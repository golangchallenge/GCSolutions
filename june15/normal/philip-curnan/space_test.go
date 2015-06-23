package main

import (
	"testing"
)

// input and test outputs for cutSpace
type testSpaces struct {
	in    space
	cut   space
	fit   bool
	out   [3]space
	valid [3]bool
}

// make a table of inputs to loop through
var tspaces = []testSpaces{
	// full space, no leftovers
	{in: space{x: 0, y: 0, w: 4, l: 4}, cut: space{x: 0, y: 0, w: 4, l: 4}, fit: true,
		out:   [3]space{{x: 0, y: 0, w: 4, l: 4}, {x: 0, y: 0, w: 4, l: 4}, {x: 0, y: 0, w: 4, l: 4}},
		valid: [3]bool{true, false, false}},
	// too big in width
	{in: space{x: 1, y: 1, w: 3, l: 3}, cut: space{x: 0, y: 0, w: 4, l: 3}, fit: false,
		out:   [3]space{{x: 0, y: 0, w: 0, l: 0}, {x: 0, y: 0, w: 0, l: 0}, {x: 0, y: 0, w: 0, l: 0}},
		valid: [3]bool{false, false, false}},
	// top row only
	{in: space{x: 0, y: 3, w: 1, l: 4}, cut: space{x: 0, y: 3, w: 1, l: 4}, fit: true,
		out:   [3]space{{x: 0, y: 3, w: 1, l: 4}, {x: 0, y: 0, w: 0, l: 0}, {x: 0, y: 0, w: 0, l: 0}},
		valid: [3]bool{true, false, false}},
	// too big in length
	{in: space{x: 3, y: 0, w: 4, l: 1}, cut: space{x: 0, y: 0, w: 1, l: 2}, fit: false,
		out:   [3]space{{x: 0, y: 0, w: 0, l: 0}, {x: 0, y: 0, w: 0, l: 0}, {x: 0, y: 0, w: 0, l: 0}},
		valid: [3]bool{false, false, false}},
	// right col only
	{in: space{x: 0, y: 3, w: 1, l: 4}, cut: space{x: 0, y: 0, w: 1, l: 4}, fit: true,
		out:   [3]space{{x: 0, y: 3, w: 1, l: 4}, {x: 0, y: 0, w: 0, l: 0}, {x: 0, y: 0, w: 0, l: 0}},
		valid: [3]bool{true, false, false}},
	// both remnants valid
	{in: space{x: 1, y: 1, w: 3, l: 3}, cut: space{x: 0, y: 0, w: 2, l: 2}, fit: true,
		out:   [3]space{{x: 1, y: 1, w: 2, l: 2}, {x: 1, y: 3, w: 1, l: 3}, {x: 3, y: 1, w: 2, l: 1}},
		valid: [3]bool{true, true, true}},
	// halfzies
	{in: space{x: 0, y: 0, w: 4, l: 4}, cut: space{x: 0, y: 0, w: 2, l: 2}, fit: true,
		out:   [3]space{{x: 0, y: 0, w: 2, l: 2}, {x: 0, y: 2, w: 2, l: 4}, {x: 2, y: 0, w: 2, l: 2}},
		valid: [3]bool{true, true, true}},
}

func TestCutSpace(t *testing.T) {
	for index, ts := range tspaces {
		out, fit, v := ts.in.carveSpace(ts.cut)
		// does it fit as expected
		if fit != ts.fit {
			t.Fatalf("test row %d fits %v, expected fits %v", index, fit, ts.fit)
		} else if !fit {
			continue
		}
		// make sure each output is valid or invalid as expected
		// if valid then check out space matches expected
		for n := 0; n < 3; n++ {
			if v[n] != ts.valid[n] {
				t.Fatalf("test row %d ts.out[%d].isValid() = %v, expected %v", index, n, ts.out[n].isValid(), ts.valid[n])
			} else if !ts.valid[n] {
				break
			} else if out[n] != ts.out[n] {
				t.Fatalf("test row %d out[%d] = %v, expected %v", index, n, out[n], ts.out[n])
			}
		}
	}
}
