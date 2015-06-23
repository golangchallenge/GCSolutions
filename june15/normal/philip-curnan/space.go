package main

import (
	"fmt"
	"os"
)

// space is the representation(s) of various areas of a pallet
type space struct {
	x, y uint8 // where is the space located
	w, l uint8 // how wide and long is it
}

// isValid returns true if the space has proper location and size
func (s space) isValid() bool {
	defer func() {
		if x := recover(); x != nil {
			fmt.Printf("panic in space.isValid: error: %v\n", x)
			os.Exit(1)
		}
	}()
	return checkLocation(s.x, s.y) && checkBounds(s.w, s.l)
}

// carveSpace removes input space 'cut' from the lower left corner of s.
// Return cut space (out[0]) and remaining spaces above (out[1]) and
// to the right (out[2]), and results of validty tests.
func (s space) carveSpace(cut space) (out []space, fits bool, valid []bool) {
	defer func() {
		if x := recover(); x != nil {
			fmt.Printf("panic s carveSpace: error: %v\n", x)
			os.Exit(1)
		}
	}()
	out = make([]space, 3)
	valid = make([]bool, 3)
	// if 'cut' is bigger then 's' or s lower left + 'cut' size bigger than pallet, it doesn't fit
	if cut.w > s.w || cut.l > s.l || s.y+cut.w > palletWidth || s.x+cut.l > palletLength {
		return out, false, valid
	}
	// remove 'cut' from the lower left of 's'
	out[0] = space{x: s.x, y: s.y, w: cut.w, l: cut.l}
	// duplicate some code for performance, cuts down calls to checkBounds and checkLocation
	valid[0] = out[0].x >= 0 && out[0].x < palletLength && out[0].y >= 0 && out[0].y < palletWidth &&
		out[0].w > 0 && out[0].w <= palletWidth && out[0].l > 0 && out[0].l <= palletLength

	// calculate what part of 's' remains above the part cut out
	out[1] = space{x: s.x, y: s.y + cut.w, w: s.w - cut.w, l: s.l}
	valid[1] = out[1].x >= 0 && out[1].x < palletLength && out[1].y >= 0 && out[1].y < palletWidth &&
		out[1].w > 0 && out[1].w <= palletWidth && out[1].l > 0 && out[1].l <= palletLength

	// calculate what part of 's' remains to the right of the part cut out
	out[2] = space{x: s.x + cut.l, y: s.y, w: cut.w, l: s.l - cut.l}
	valid[2] = out[2].x >= 0 && out[2].x < palletLength && out[2].y >= 0 && out[2].y < palletWidth &&
		out[2].w > 0 && out[2].w <= palletWidth && out[2].l > 0 && out[2].l <= palletLength
	return out, true, valid
}

// checkBounds returns true if the given width and length will fit on a pallet
func checkBounds(w, l uint8) bool {
	return w > 0 && w <= palletWidth && l > 0 && l <= palletLength
}

// checkLocation returns true if the given coordinates are within a pallet grid
func checkLocation(x, y uint8) bool {
	return x >= 0 && x < palletLength && y >= 0 && y < palletWidth
}
