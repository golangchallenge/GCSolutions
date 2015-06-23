//
// =========================================================================
//
//       Filename:  grid.go
//
//    Description:  Implements a grid by using slices not lists.
//
//           TODO:  Try to get it thread safe. Resources:
//      						https://github.com/hishboy/gocommons/blob/master/lang/stack.go
//                  https://gist.github.com/moräs/2141121
//
//
//        License:  GNU General Public License
//      Copyright:  Copyright (c) 2015, Frank Milde
//
// =========================================================================
//

package main

import (
	"fmt"
	"log"
	"sort"
)

type Orientation uint8

const (
	HORIZONTAL Orientation = iota
	VERTICAL
	SQUAREGRID
)

type GridElement struct {
	x, y   uint8       //origin
	w, l   uint8       //width length
	size   int         // size
	orient Orientation //horizontal, vertical, square
}

type FreeGrid []GridElement

var emptygrid = GridElement{}

func NewGrid() FreeGrid {
	var g []GridElement
	return g
}
func NewInitialGrid() FreeGrid {
	init := GridElement{0, 0, 4, 4, 16, SQUAREGRID}
	f := []GridElement{init}
	return f
}
func NewSubGrid(g GridElement) FreeGrid {
	f := []GridElement{g}
	return f
}

func (e *GridElement) SetProperties() {

	e.size = int(e.l * e.w)

	if e.l == e.w {
		e.orient = SQUAREGRID
	}

	if e.w > e.l {
		e.orient = HORIZONTAL
	}
	if e.w < e.l {
		e.orient = VERTICAL
	}
}

func (g FreeGrid) IsEmpty() bool { return len(g) == 0 }

func (orient Orientation) String() string {

	var s string

	switch orient {
	case HORIZONTAL:
		s = "horizontal"
	case VERTICAL:
		s = "vertical"
	case SQUAREGRID:
		s = "square"
	}

	return s
}
func (e GridElement) String() string {

	var s string
	s += fmt.Sprintf("[%d %d %d %d] ", e.x, e.y, e.w, e.l)
	s += fmt.Sprintf("%d %v ", e.size, e.orient)
	return s
}
func (g FreeGrid) String() string {

	var s string
	for i, g := range g {
		boxtmp := box{g.x, g.y, g.w, g.l, 1}
		grid := pallet{[]box{boxtmp}}
		if i < 10 {
			s += fmt.Sprintf("[ %d]   -->   %v,%v\n", i, g, grid)
		} else {
			s += fmt.Sprintf("[%d]   -->   %v,%v\n", i, g, grid)
		}
	}
	return s
}

// Put takes a box b and puts it in the lower left corner of Gridelement e.
// If b does not cover e completely, the remaining free space of grid e is
// returned. This return value is of type FreeGrid := []GridElement and
// contains up to three elements into which the original e has been
// split by the box: (1) top, (2) right, (3) top right
//  | 1 1 1 3 |
//  | 1 1 1 3 |
//  | b b b 2 |
//  | b b b 2 |
func Put(b *box, e GridElement) FreeGrid {

	errCoor := b.SetOrigin(e.x, e.y)
	if errCoor != nil {
		log.Println("Error when setting origin ", e.x, e.y, " of box ", b.l, b.w, b.id)
		log.Println(e)
		log.Println(b)
	}

	bottom := GridElement{
		x: b.x + b.l,
		y: b.y,
		w: b.w,
		l: e.l - b.l,
	}
	right := GridElement{
		x: b.x,
		y: b.y + b.w,
		w: e.w - b.w,
		l: b.l,
	}
	bottomRight := GridElement{
		x: b.x + b.l,
		y: b.y + b.w,
		w: e.w - b.w,
		l: e.l - b.l,
	}
	bottom.SetProperties()
	right.SetProperties()
	bottomRight.SetProperties()

	elements := []GridElement{bottom, right, bottomRight}

	var split FreeGrid

	for _, e := range elements {
		if e.size != 0 {
			split = append(split, e)
		}
	}

	sort.Sort(ByArea(split))

	return split
}

// Update cuts last element of Freegrid f and replaces it with a new
// FreeGrid newG.
func (f *FreeGrid) Update(newG FreeGrid) {

	//	Cut last element
	last := len(*f) - 1
	(*f) = (*f)[:last]

	//	Append new FreeGrid
	if !newG.IsEmpty() {
		*f = append(*f, newG...)
	}
}

//
//
//
// GridElementsAreEqual compares each field of two GridElements a,b and
// return true if they are equal.
func GridElementsAreEqual(a, b GridElement) bool {
	if a.size != b.size {
		return false
	}
	if a.orient != b.orient {
		return false
	}
	if a.x != b.x {
		return false
	}
	if a.y != b.y {
		return false
	}
	if a.l != b.l {
		return false
	}
	if a.w != b.w {
		return false
	}

	return true
} // -----  end of function GridElementsAreEqual  -----

// FreeGridsAreEqual compares all GridElements of two FreeGrids  a,b and
// return true if they are equal.
func FreeGridsAreEqual(a, b FreeGrid) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if !GridElementsAreEqual(v, b[i]) {
			return false
		}
	}
	return true
} // -----  end of function FreeGridssAreEqual  -----

// =========================================================================
//  Implementing Sort interface
//  Will order boxes from lowest to highest size.
//  Use as:
//          boxes = []box
//          sort.Sort(BySize(boxes))
//
//	  			box{0, 0, 4, 4, 101},       box{0, 0, 2, 1, 103},
//	  			box{0, 0, 2, 2, 102},  -->  box{0, 0, 2, 2, 102},
//	  			box{0, 0, 2, 1, 103},       box{0, 0, 3, 2, 104},
//	  			box{0, 0, 3, 2, 104},       box{0, 0, 4, 4, 101},
// =========================================================================
type ByArea []GridElement

func (a ByArea) Len() int           { return len(a) }
func (a ByArea) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByArea) Less(i, j int) bool { return a[i].size < a[j].size }

// -----  end of Sort Interface  -----
