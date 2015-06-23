package main

//We won't store box types instead we will store
//the total number of certain size of boxes and ids.
type boxInfo struct {
	count int
	ids   []uint32
}

//box with 2 width and 3 length is in basement[3][2] cell.
type boxStorage struct {
	basement [palletWidth][palletLength]boxInfo
}

//isEmpty checks if boxstorage doesn't contain any boxes.
func (bs *boxStorage) isEmpty() bool {
	//our boxStorage looks like right triangle
	//since we applied canon function to every boxes
	for w := 0; w < palletWidth; w++ {
		for y := 0; y <= w; y++ {
			if bs.basement[y][w].count > 0 {
				return false
			}
		}
	}
	return true
}

func (bs *boxStorage) addBox(b box) {
	// we canon boxes to easily sort them later.
	b = b.canon()
	bs.basement[b.l-1][b.w-1].count++
	bs.basement[b.l-1][b.w-1].ids = append(bs.basement[b.l-1][b.w-1].ids, b.id)
}

//getBiggestBox finds biggest possible box which doesn't
//have greater width than maxWidth and length than maxLength.
func (bs *boxStorage) findBiggestBox(maxWidth, maxLength uint8) (b box) {
	shouldRotate := false

	//normally we find boxes for spaces with width that is greater than length
	//but if the biggest space has opposite values, we temporarily rotate
	//to our normal state and revert the result back.
	if maxWidth < maxLength {
		shouldRotate = true
		maxWidth, maxLength = maxLength, maxWidth
	}

	for w := int(maxWidth) - 1; w >= 0; w-- {
		for l := int(maxLength) - 1; l >= 0; l-- {
			//a box is found
			if bs.basement[l][w].count > 0 {
				bsCount := bs.basement[l][w].count
				boxID := bs.basement[l][w].ids[bsCount-1]
				bs.basement[l][w].ids = bs.basement[l][w].ids[:bsCount-1]
				bs.basement[l][w].count--
				if shouldRotate {
					b = box{0, 0, uint8(l + 1), uint8(w + 1), boxID}
				} else {
					b = box{0, 0, uint8(w + 1), uint8(l + 1), boxID}
				}
				return
			}
		}
	}
	//no suitable box found
	b = box{0, 0, 0, 0, 0}
	return
}

type freeSpace struct {
	x      uint8
	y      uint8
	width  uint8
	length uint8
}

func (fs *freeSpace) area() uint8 {
	return fs.width * fs.length
}

type palletStorage struct {
	complete bool
	pallet
}

//isFull checks whether pallet has any free space
func (ps *palletStorage) isFull() bool {
	var sum uint8
	sum = 0
	for _, x := range ps.boxes {
		sum += x.l * x.w
	}
	if sum == palletLength*palletWidth {
		return true
	}
	return false
}

//findBiggestSpace finds biggest space that doesn't contain any boxes.
func (ps *palletStorage) findBiggestSpace() freeSpace {
	var x, y uint8
	var tempSpace freeSpace
	biggestSpace := freeSpace{0, 0, 0, 0}

	for y = 0; y < palletLength; y++ {
		for x = 0; x < palletWidth; x++ {
			if ps.isFreeSpace(int(x), int(y)) {
				//box length will be from y to maxY.
				maxY := y
				for maxY < palletLength && ps.isFreeSpace(int(x), int(maxY)) {
					maxY++
				}
				maxY--
				rightCount := ps.rightFreeCells(x, y, maxY)
				leftCount := ps.leftFreeCells(x, y, maxY)
				tempSpace = freeSpace{x - leftCount, y, leftCount + rightCount + 1, maxY - y + 1}
				if tempSpace.area() > biggestSpace.area() {
					biggestSpace = tempSpace
				}
			}
		}
	}
	return biggestSpace
}

//rightFreeCells counts free adjacant cells with same height maxY
//on the right side of the given cell.
func (ps *palletStorage) rightFreeCells(x, y, maxY uint8) (count uint8) {
	count = 0
	tempX := x + 1
	for tempX < palletWidth {
		//checking whether any boxes already placed on cells from y to maxY.
		for tempY := y; tempY <= maxY; tempY++ {
			if !ps.isFreeSpace(int(tempX), int(tempY)) {
				return
			}
		}
		count++
		tempX++
	}
	return
}

//leftFreeCells counts free adjacant cells with same height maxY
//on the left side of the given cell.
func (ps *palletStorage) leftFreeCells(x, y, maxY uint8) (count uint8) {
	count = 0
	tempX := int(x) - 1
	for tempX >= 0 {
		for tempY := y; tempY <= maxY; tempY++ {
			if !ps.isFreeSpace(tempX, int(tempY)) {
				return
			}
		}
		count++
		tempX--
	}
	return
}

func (ps *palletStorage) isFreeSpace(x, y int) bool {
	for _, b := range ps.boxes {
		//since we used cartesian coordinate system and driver program uses coordinate system
		//that is exactly opposite of our system, we swap b.x and b.y to check free space.
		if x < int(b.y)+int(b.w) && x >= int(b.y) && y < int(b.x)+int(b.l) && y >= int(b.x) {
			return false
		}
	}
	return true
}
