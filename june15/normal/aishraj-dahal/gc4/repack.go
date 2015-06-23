package main

//This implementation is based on http://clb.demon.fi/files/RectangleBinPack.pdf

// A repacker repacks trucks.
type repacker struct {
}

type fit int

type shelf struct {
	minHeight, maxHeight, width uint8
}

//
const (
	_               = iota
	verticalFit fit = 1 << iota
	horizontalFit
	newShelfFit
	unFit
)

func (f fit) String() string {
	switch f {
	case verticalFit:
		return "VerticalFit"
	case horizontalFit:
		return "horizontalFit"
	case newShelfFit:
		return "newShelfFit"
	case unFit:
		return "unFit"
	default:
		return "Unknown Value"
	}
}

// shelfNF uses a 'shelf' based implemeantion to pack the given pallets.
// It accepts a truck, collects all boxes , packs them and put them  back in a truck
// and returns the truck.
func shelfNF(t *truck) (out *truck) {
	out = &truck{id: t.id}
	var boxes []box
	//collect all boxes
	for _, p := range t.pallets {
		for _, b := range p.boxes {
			boxes = append(boxes, b)
		}
	}
	var outPallets []pallet
	var uno pallet
	var shelves []shelf

	for _, item := range boxes {
		if len(shelves) == 0 {
			uno, shelves = makePallet(item)
			continue
		}
		fitness := findFit(item, &shelves)
		switch fitness {
		case horizontalFit:
			item = sideWays(item)
			uno = addToPallet(uno, item, &shelves)
			break
		case verticalFit:
			item = upRight(item)
			uno = addToPallet(uno, item, &shelves)
			break
		case newShelfFit:
			topShelf := shelves[len(shelves)-1]
			item = sideWays(item)
			newTopShelf := shelf{minHeight: topShelf.maxHeight, maxHeight: (item.l + topShelf.maxHeight), width: 0}
			shelves = append(shelves, newTopShelf)
			uno = addToPallet(uno, item, &shelves)
			break
		case unFit:
			//this means that uno is out of space.
			//we add it to the staging area and get a new pallet
			dummyTruck := truck{id: 0}
			dummyTruck.pallets = outPallets
			outPallets = append(outPallets, uno)
			uno, shelves = makePallet(item)
			break
		default:
			panic("Does not fit anywhere, despite allocating a new pallet. Its an error")
		}
	}
	if len(uno.boxes) > 0 {
		outPallets = append(outPallets, uno)
	}
	out.pallets = outPallets
	return
}

// findFit finds out whether the box fits the given shelf or not.
// In case it does fit the shelf, it returns the orienatation of the fitness.
func findFit(item box, shelves *[]shelf) fit {
	topShelf := (*shelves)[len(*shelves)-1]
	upBox := upRight(item)
	sideBox := sideWays(item)
	if upBox.l+topShelf.minHeight <= topShelf.maxHeight && (upBox.w+topShelf.width <= palletWidth) && (upBox.l+topShelf.minHeight <= palletLength) {
		return verticalFit
	}
	if sideBox.l+topShelf.minHeight <= topShelf.maxHeight && (sideBox.w+topShelf.width <= palletWidth) && (sideBox.l+topShelf.minHeight <= palletLength) {
		return horizontalFit
	}
	if (sideBox.l + topShelf.maxHeight) <= palletLength {
		return newShelfFit
	}
	return unFit
}

func addToPallet(uno pallet, item box, shelves *[]shelf) pallet {
	currentTop := (*shelves)[len(*shelves)-1]
	item.x = currentTop.minHeight
	item.y = currentTop.width
	currentTop.width += item.w
	(*shelves)[len(*shelves)-1] = currentTop
	uno.boxes = append(uno.boxes, item)
	return uno
}

func makePallet(item box) (packet pallet, shelves []shelf) {
	item = item.canon()

	bottomShelf := shelf{minHeight: 0, maxHeight: item.l, width: item.w}
	shelves = append(shelves, bottomShelf)
	packet = pallet{boxes: []box{item}}
	return
}

func newRepacker(in <-chan *truck, out chan<- *truck) *repacker {
	go func() {
		for t := range in {
			// The last truck is indicated by its id. You might
			// need to do something special here to make sure you
			// send all the boxes.
			if t.id == idLastTruck {
			}

			//t = oneBoxPerPallet(t)
			t = shelfNF(t)
			out <- t
		}
		// The repacker must close channel out after it detects that
		// channel in is closed so that the driver program will finish
		// and print the stats.
		close(out)
	}()
	return &repacker{}
}
