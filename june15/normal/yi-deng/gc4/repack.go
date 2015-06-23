package main

import (
	"sync"
)

type repacker struct {
}

// The area of a pallet
const palletArea = palletWidth * palletLength

// Extra info about a pallet.
type palletInfo struct {
	// The index of pallet in the truck
	indexInTruck int
	// occupied[x][y] is the number of occupied area in the left-top part of (x,y), inclusive.
	occupied [palletLength + 1][palletWidth + 1]uint8
	// max possible fit box width/length. They are only larger than extual values.
	maxBoxW, maxBoxL uint8
}

// putAt marks the box in the pallet as occupied. Updates maxBoxL/W accordingly.
// @return the free space
func (info *palletInfo) putAt(x, y, w, l uint8) uint8 {
	// Update info.occupied
	for i := y + 1; i <= palletLength; i++ {
		ww := i - y
		if ww > w {
			ww = w
		}
		for j := x + 1; j <= palletWidth; j++ {
			ll := j - x
			if ll > l {
				ll = l
			}
			info.occupied[i][j] += ww * ll
		}
	}

	if w == palletWidth {
		// a full-width box reduces maxBoxL by l
		info.maxBoxL -= l
	} else if l == palletLength {
		// a full-length box reduces maxBoxW by w
		info.maxBoxW -= w
	}

	return palletArea - info.occupied[palletLength][palletWidth]
}

// fitBox finds a position for a box of specified w/l
func (info *palletInfo) fitBox(w, l uint8) (x, y uint8, succ bool) {
	for y := uint8(0); y <= palletLength-w; y++ {
		for x := uint8(0); x <= palletWidth-l; x++ {
			//   +(x,y)|  |
			// --------+--+ -(x+l,y)
			//         |  |
			// --------+--+
			//  -(x,y+w)    +(x+l,y+w)
			if info.occupied[y+w][x+l]-info.occupied[y+w][x]-info.occupied[y][x+l]+info.occupied[y][x] == 0 {
				return x, y, true
			}
		}
	}

	return 0, 0, false
}

// Array of palletInfo slices by free spaces.
type palletInfos [palletArea][]*palletInfo

// placeBox places a box in a truck. It first tries to find a pallet that can fit the box. If failed, a
// new pallet is created.
func placeBox(infos *palletInfos, area, w, l uint8, id uint32, t *truck) {
	// Since we are placing boxes from larger to smaller, not possible to have pallets with free spaces
	// larger than palletArea - area.
	for freeSpace := palletArea - area; freeSpace >= area; freeSpace-- {
		for i, info := range infos[freeSpace] {
			if w > info.maxBoxW || l > info.maxBoxL {
				// Not possible to fit.
				continue
			}

			if x, y, succ := info.fitBox(w, l); succ {
				// Put the box to the pallet
				newFreeSpace := info.putAt(x, y, w, l)
				t.pallets[info.indexInTruck].boxes = append(t.pallets[info.indexInTruck].boxes, box{
					x:  x,
					y:  y,
					l:  l,
					w:  w,
					id: id,
				})

				// Remove from the current list
				n1 := len(infos[freeSpace]) - 1
				infos[freeSpace][i] = infos[freeSpace][n1]
				infos[freeSpace] = infos[freeSpace][:n1]

				if newFreeSpace > 0 {
					// The following updating of maxBoxL/W depends on the fact that we allways try place
					// a box to the left/top position of the pallet. If the box touch right/bottom
					// edge, a lx1 or 1xw box should not fit this pallet anymore.
					if x+l == palletLength && info.maxBoxW == palletWidth {
						info.maxBoxW = palletWidth - 1
					}
					if y+w == palletWidth && info.maxBoxL == palletLength {
						info.maxBoxL = palletLength - 1
					}

					// Append info to the new list
					infos[newFreeSpace] = append(infos[newFreeSpace], info)
				}
				return
			}

			// not fit, try update maxBoxW/L

			if l == 1 && w <= info.maxBoxW {
				// update maxBoxW
				info.maxBoxW = w - 1
			}

			if w == 1 && l <= info.maxBoxL {
				// update maxBoxL
				info.maxBoxL = l - 1
			}
		}
	}

	// No existing pallet fit, require a new one
	info := &palletInfo{
		indexInTruck: len(t.pallets),
		maxBoxW:      palletWidth,
		maxBoxL:      palletLength,
	}
	// Put the box in the pallet.
	newFreeSpace := info.putAt(0, 0, w, l)

	// Put info into infos by its free space.
	infos[newFreeSpace] = append(infos[newFreeSpace], info)

	// Place the box as the first box in the new pallet
	t.pallets = append(t.pallets, pallet{boxes: []box{box{
		w:  w,
		l:  l,
		id: id,
	}}})
}

// repackTruck repacks pallets in a single truck.
func repackTruck(t *truck) *truck {
	// Collect boxes by area.
	var boxesByArea [palletArea + 1][]box
	for _, p := range t.pallets {
		for _, b := range p.boxes {
			area := b.w * b.l
			boxesByArea[area] = append(boxesByArea[area], b)
		}
	}

	// Place boxes same size as the pallet
	t.pallets = make([]pallet, 0, len(boxesByArea[palletArea]))
	for _, b := range boxesByArea[palletArea] {
		t.pallets = append(t.pallets, pallet{boxes: []box{box{
			w:  palletWidth,
			l:  palletLength,
			id: b.id,
		}}})
	}

	var infos palletInfos
	// Place boxes from large to small
	for area := uint8(palletArea - 1); area > 0; area-- {
		for _, b := range boxesByArea[area] {
			placeBox(&infos, area, b.w, b.l, b.id, t)
		}
	}

	return t
}

func newRepacker(in <-chan *truck, out chan<- *truck) *repacker {
	go func() {
		var wg sync.WaitGroup
		var onetruck *truck
		for t := range in {
			// All pallets are collected to a single truck for repacking. Trucks other than the first one
			// will be emptied. In practice, this can be easily replaced by a distribution
			// of repacked pallets evenly to all trucks. But since the challenge did not define
			// this part in details, we just use a simplest way.
			if onetruck == nil {
				onetruck = t
			} else {
				onetruck.pallets = append(onetruck.pallets, t.pallets...)
				t.pallets = nil
				out <- t
			}

			if t.id == idLastTruck || len(onetruck.pallets) > 4000 {
				// Starts a new goroutine for every 1000 pallets.
				wg.Add(1)
				go func(t *truck) {
					defer wg.Done()
					out <- repackTruck(t)
				}(onetruck)
				onetruck = nil
			}
		}
		if onetruck != nil {
			out <- repackTruck(onetruck)
			onetruck = nil
		}
		wg.Wait()
		close(out)
	}()
	return &repacker{}
}
