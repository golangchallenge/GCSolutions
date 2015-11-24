package main

import "fmt"

//Binarymask* use in binary operations
const (
	BinaryMask9  = 511 // 111111111
	BinaryMask3  = 7   // 111
	BinaryMask12 = 73  // 1001001
)

//Binary gathered basic methods with binary mask
type Binary struct{}

//ConvBit2Slice return slice with index of bytes which equal 1
func (Binary) ConvBit2Slice(mask int) (pos []int) {
	i := 0
	for mask > 0 {
		if 1 == mask&0x01 {
			pos = append(pos, i)
		}
		i++
		mask = mask >> 1
	}
	return
}

//InvertBites work only for 9 bits
func (Binary) InvertBites(mask int) (invertedMask int) {
	return mask ^ BinaryMask9
}

//CountZero calculate zero values in mask (check all 9 bits)
func (Binary) CountZero(mask int) (cnt int) {
	if 0 == mask {
		return 9 //For the input 0 it should output 9
	}
	for mask > 0 {
		//easy to count "1" and then substract it from 9
		if 1 == mask&0x01 {
			cnt++
		}
		mask >>= 1
	}
	cnt = 9 - cnt
	return cnt
}

//CheckBit return true if in position `pos` set `1`
func (Binary) CheckBit(mask, pos int) bool {
	return mask>>uint(pos)&0x01 == 1
}

//Box hepter to transform coordinates fron box to board
type Box struct{}

//BoxID return boxID by cell coordinates
func (Box) BoxID(x, y int) int {
	return (y/3)*3 + (x / 3)
}

//FirstXYBox return upper-left cell's coordinates for box with boxID
func (Box) FirstXYBox(boxID int) (int, int) {
	return (boxID % 3) * 3, (boxID / 3) * 3
}

//XYBox return upper-left cell's coordinates for box with boxID
func (Box) XYBox(boxID, posInBox int) (int, int) {
	return (boxID%3)*3 + posInBox%3, (boxID/3)*3 + posInBox/3
}

//BoxMinFreeCell find the box with less count of free cells
//iterate though each box on the board
func (b Box) BoxMinFreeCell(maskSet [9]int) (boxID int) {
	minCnt := 9
	bin := &Binary{}
	for b, boxMask := range maskSet {
		zeroCnt := bin.CountZero(boxMask)
		if zeroCnt > 0 && zeroCnt < minCnt {
			minCnt = zeroCnt
			boxID = b
		}
	}
	return
}

//Cell - the smallest structure in this quiz
type Cell struct {
	x, y, v int
}

//Number contains statistics info about each number and it positions on the board
type Number struct {
	cnt, r, c, b int
}

func (n Number) String() string {
	return fmt.Sprintf("(%d) rows: %d (%b), columns: %d (%b); boxes: %d (%b)\n", n.cnt, n.r, n.r, n.c, n.c, n.b, n.b)
}
