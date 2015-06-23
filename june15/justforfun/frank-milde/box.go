//
// =========================================================================
//
//       Filename:  box.go
//
//    Description:  Handles all things related to boxes.
//
//        License:  GNU General Public License
//      Copyright:  Copyright (c) 2015, Frank Milde
//
// =========================================================================
//

package main

import (
	"errors"
	//	"fmt"
)

// ===  FUNCTION  ==========================================================
//         Name:  HasValidDimensions
//  Description:  Checks if a box is
//                - small enough to fit on an empty pallet.
//                - has a non zero length and width
// =========================================================================
func (b *box) HasValidDimensions() bool {
	return (b.l <= palletWidth) && (b.w <= palletLength) && (b.l > 0) && (b.w > 0)
} // -----  end of function HasValidDimensions  -----

// ===  FUNCTION  ==========================================================
//         Name:  ValidCoordinates
//  Description:  Checks if x,y coordinates are within pallet bounds.
// =========================================================================
func ValidCoordinates(x, y uint8) bool {
	return (y < palletWidth) && (x < palletLength)
} // -----  end of function ValidCoordinates  -----

// ===  FUNCTION  ==========================================================
//         Name:  HasValidCoordinates
//  Description:  Checks if the origin of a box is within pallet bounds.
// =========================================================================
func (b *box) HasValidCoordinates() bool {
	return ValidCoordinates(b.x, b.y)
} // -----  end of function HasValidCoordinates  -----

// ===  FUNCTION  ==========================================================
//         Name:  IsWithinBounds
//  Description:  Checks if a box fits within the pallet bounds.
// =========================================================================
func (b *box) IsWithinBounds(x, y uint8) bool {
	boxIsTooWide := (b.l + x) > palletWidth
	boxIsTooLong := (b.w + y) > palletLength
	return (!boxIsTooWide && !boxIsTooLong)
} // -----  end of function IsWithinBounds  -----

func (b *box) Size() uint8 { return b.l * b.w }
func (b *box) Rotate() {
	tmp := b.l
	b.l = b.w
	b.w = tmp
} // -----  end of function Rotate  -----
func (b *box) IsSquare() bool {
	return b.l == b.w
}
func (b *box) Display() string {
	c := b.canon()

	var out string
	var i, j uint8

	for i = 0; i < c.w; i++ {
		for j = 0; j < c.l; j++ {
			out += "x "
		}
		out += "\n"
	}
	return out
}

// ===  FUNCTION  ==========================================================
//         Name:  BoxesAreEqual
//  Description:  Compares if two Boxes are equal. Since we cannot simply
//                range over structs we have do it manually. The input is as
//                a value, not pointer to use this method in
//                `PalletsAreEqual` as we cannot range over pointers
//         TODO:  Rewrite to use pointers instead of values.
// =========================================================================
func BoxesAreEqual(a, b box) bool {
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
	if a.id != b.id {
		return false
	}

	return true

	// It would be more elegant to use reflections to get the values of the
	// respected fields of the box struct. See also:
	// https://stackoverflow.com/qüstions/18926303/iterate-through-a-struct-in-go
	//
	// However, using reflect to iterate over the box structure fails as the
	// data field variables are all lower case in `box` and thus are invisible
	// outside the defining package and reflect is an outside package. See
	// https://groups.google.com/forum/#!topic/golang-nuts/UYgse9hnfoc
	//
	//
	//	A := reflect.ValueOf(a)
	//	B := reflect.ValueOf(b)
	//
	//	A_values := make([]interface{}, A.NumField())
	//	B_values := make([]interface{}, B.NumField())
	//
	//	for i := 0; i < A.NumField(); i++ {
	//		A_values[i] = A.Field(i).Interface()
	//	}
	//	for i := 0; i < B.NumField(); i++ {
	//		B_values[i] = B.Field(i).Interface()
	//	}
	//
	//	for i, v := range A_values {
	//		if v != B_values[i] {
	//			return false
	//		}
	//	}
	//	return true
} // -----  end of function BoxesAreEqual  -----
func BoxArraysAreEqual(a, b []box) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if !BoxesAreEqual(v, b[i]) {
			return false
		}
	}
	return true
} // -----  end of function BoxArraysAreEqual  -----
func PalletsAreEqual(a, b pallet) bool {
	if len(a.boxes) != len(b.boxes) {
		return false
	}
	for i, v := range a.boxes {
		if !BoxesAreEqual(v, b.boxes[i]) {
			return false
		}
	}
	return true
} // -----  end of function PalletssAreEqual  -----

func (b box) AddToPallet(p *pallet) {
	//	fmt.Println("In box: ", b)
	if BoxesAreEqual(b, emptybox) {
		return
	}
	if !b.HasValidCoordinates() {
		return
	}

	p.boxes = append(p.boxes, b)
	//	fmt.Println("p.boxes: ", p.boxes)
} // -----  end of function AddToPallet  -----

// ===  FUNCTION  ==========================================================
//         Name:  SetOrigin
//  Description:  Places Box on Grid. Returns Error when failed.
// =========================================================================
func (b *box) SetOrigin(x, y uint8) error {
	if !ValidCoordinates(x, y) {
		return errors.New("box: Origin coordinates out of bounds.")
	}
	if !b.HasValidDimensions() {
		return errors.New("box: Has invalid size.")
	}
	if b.IsWithinBounds(x, y) {
		b.x = x
		b.y = y
		return nil
	} else {
		b.Rotate()
		if b.IsWithinBounds(x, y) {
			b.x = x
			b.y = y
			return nil
		} else {
			return errors.New("box.SetOrigin: Hangs over pallet edge. Unable to place box on grid")
		}
		return errors.New("box.SetOrigin: Hangs over pallet edge. Unable to place box on grid")
	}
} // -----  end of function SetOrigin  -----

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
type BySize []box

func (a BySize) Len() int           { return len(a) }
func (a BySize) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a BySize) Less(i, j int) bool { return a[i].Size() < a[j].Size() }

// -----  end of Sort Interface  -----
