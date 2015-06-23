package main

import (
	//"fmt"
	"math"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
)

// A repacker repacks trucks.
type repacker struct {
	boxes            []box
	lowerBound       uint
	pallets          []pallet
	palletsFreeSpace []uint
}

// ByArea used to sort boxes by area
type ByArea []box

func min(i, j uint8) int {
	if i < j {
		return int(i)
	}
	return int(j)
}

func (b ByArea) Len() int      { return len(b) }
func (b ByArea) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b ByArea) Less(i, j int) bool {
	iArea := b[i].w * b[i].l
	jArea := b[j].w * b[j].l

	if iArea == jArea {
		return min(b[i].w, b[i].l) > min(b[j].w, b[j].l)
	}

	return iArea > jArea
}

var emptyPalletFreeSpace uint = palletLength * palletWidth

type boxPlacement struct {
	s  float64
	b  box
	pn int
}

// packs pallets using modified Touching Perimeter algorithm
func packBoxesToPallets(boxes []box) (out []pallet) {
	// sort box container
	sort.Sort(ByArea(boxes))

	// canonize & horizontaly orient box container
	for i, b := range boxes {
		boxes[i] = b.canon()
	}
	//fmt.Println("Sorted boxes: ", boxes)

	// calculate lower bound (sum(box[i].w*box[i].l) / emptyPalletFreeSpace)
	lowerBound := uint(0)
	for _, b := range boxes {
		lowerBound += uint(b.l * b.w)
	}
	lowerBound = uint(math.Ceil(float64(lowerBound) / float64(emptyPalletFreeSpace)))
	//fmt.Println("Lower bound: ", lowerBound)

	pallets := make([]pallet, lowerBound)
	palletsFreeSpace := make([]uint, lowerBound)
	for i := range palletsFreeSpace {
		palletsFreeSpace[i] = emptyPalletFreeSpace
	}

	for _, b := range boxes {
		//fmt.Println("Packing box: ", b)
		maxScore := float64(0)
		maxScoreBox := box{}
		maxScorePallet := 0

		// channel for all possible box placements with score
		boxPlacements := make(chan boxPlacement)
		go func() {
			// after this function ends (all placments are found) close the channel
			defer close(boxPlacements)

			cpus := runtime.GOMAXPROCS(0)
			wg := sync.WaitGroup{}
			// there will be cpus subroutines
			wg.Add(cpus)
			// index for determining pallet for each subroutine
			idx := uint64(0)
			palletsLength := len(pallets)

			// after function ends and before closing channel, wait for all subroutines to end
			defer wg.Wait()

			// lanuch subroutines
			for n := 0; n < cpus; n++ {
				go func() {
					// after this subroutine ends, decrement WaitGroup
					defer wg.Done()
					for {
						// use idx to get pallete index, and increase it for other subroutines
						pn := int(atomic.AddUint64(&idx, uint64(1))) - 1
						// if pallete index is more than palletsLength, end this loop, to end subroutine
						if pn >= palletsLength {
							break
						}
						//fmt.Println("Try it on pallet: ", pn)
						//fmt.Println("Free space: ", palletsFreeSpace[pn])
						if palletsFreeSpace[pn] == 0 || palletsFreeSpace[pn] < uint(b.w*b.l) {
							// no free space or too low free space, box can not be placed
							continue
						}

						if palletsFreeSpace[pn] == emptyPalletFreeSpace {
							// empty pallet, justify to 0,0 and place horizontal
							boxPlacements <- boxPlacement{50.0, box{0, 0, b.w, b.l, b.id}, pn}
						} else {
							// not empty pallet, geta all posibilities for horizontal orientation
							placedBoxes := make(chan box)
							pg, _ := pallets[pn].paint()
							go getNormalPlacements(b, pg, placedBoxes)
							for bp := range placedBoxes {
								s := boxScoreOnPallet(bp, pg)
								boxPlacements <- boxPlacement{s, bp, pn}
							}
						}
					}
				}()
			}
		}()

		for bp := range boxPlacements {
			//fmt.Println("Possibility: ", bp)

			if bp.s > maxScore {
				// get top score
				maxScore = bp.s
				maxScoreBox = bp.b
				maxScorePallet = bp.pn
			} else if bp.s == maxScore {
				// score is equal, take pallet with less free space
				if palletsFreeSpace[maxScorePallet] > palletsFreeSpace[bp.pn] {
					maxScoreBox = bp.b
					maxScorePallet = bp.pn
				} else if palletsFreeSpace[maxScorePallet] == palletsFreeSpace[bp.pn] {
					// score is equal, free space is equal, take pallet with smallest index
					if maxScorePallet > bp.pn {
						maxScoreBox = bp.b
						maxScorePallet = bp.pn
					} else if maxScorePallet == bp.pn {
						// score is equal, free space is equal, pallet index is equal,
						// take horizontaly placed box (length > width)
						if maxScoreBox.l < bp.b.l {
							maxScoreBox = bp.b
						} else if maxScoreBox.l == bp.b.l {
							// score is equal, free space is equal, pallet index is equal,
							// both are horizontaly placed, take the one with lower y placement
							if maxScoreBox.y > bp.b.y {
								maxScoreBox = bp.b
							} else if maxScoreBox.y == bp.b.y {
								// score is equal, free space is equal, pallet index is equal,
								// both are horizontaly placed on the same y axis, take the lower x placement
								if maxScoreBox.x > bp.b.x {
									maxScoreBox = bp.b
								}
							}
						}
					}
				}
			}
		}
		if maxScore > 0 {
			//fmt.Println("Winnig pallet: ", maxScorePallet)
			// place winning box on winning pallet
			pallets[maxScorePallet].boxes = append(pallets[maxScorePallet].boxes, maxScoreBox)
			palletsFreeSpace[maxScorePallet] -= uint(maxScoreBox.l * maxScoreBox.w)
		} else {
			//fmt.Println("No winning pallet, creating new")
			// create new pallet and horizonatly orient box on
			pallets = append(pallets, pallet{boxes: []box{box{0, 0, b.w, b.l, b.id}}})
			palletsFreeSpace = append(palletsFreeSpace, emptyPalletFreeSpace-uint(b.l*b.w))
		}
	}

	out = []pallet{}
	for _, p := range pallets {
		if len(p.boxes) != 0 {
			out = append(out, p)
		}
	}
	return
}

// calculates free space for pallet
func getFreeSpace(p pallet) uint {
	pg, _ := p.paint()
	fs := uint(0)
	for _, b := range pg {
		if b == emptybox {
			fs++
		}
	}
	return fs
}

// sends all bossible normal box placements (bottom, left) to channel c
// after all placements are sent, closes channel
func getNormalPlacements(b box, pg palletgrid, c chan box) {
	defer close(c)
	emptyRectangles := getAllEmptyRectangles(pg)
	if len(emptyRectangles) == 0 {
		return
	}
	for _, r := range emptyRectangles {
		if r.w >= b.w && r.l >= b.l {
			c <- box{r.x, r.y, b.w, b.l, b.id}
		}
		if b.l != b.w {
			if r.w >= b.l && r.l >= b.w {
				c <- box{r.x, r.y, b.l, b.w, b.id}
			}
		}
	}
}

// gets all empty rectangles for normal box placement
func getAllEmptyRectangles(pg palletgrid) []box {
	ret := []box{}

	for i := 0; i < palletWidth; i++ {
		for j := 0; j < palletLength; j++ {
			if pg[i*palletLength+j] == emptybox {
				if i == 0 || pg[(i-1)*palletLength+j] != emptybox {
					if j == 0 || pg[i*palletLength+j-1] != emptybox {
						// is on edge, calculate rectange lenght, width
						b := box{uint8(i), uint8(j), 0, 0, 0}
						for m := i; m < palletWidth; m++ {
							if pg[m*palletLength+j] == emptybox {
								b.l++
							} else {
								break
							}
							w := uint8(1)
							for n := j + 1; n < palletLength; n++ {
								if pg[m*palletLength+n] == emptybox {
									w++
								} else {
									break
								}
							}
							if b.w == 0 {
								b.w = w
							} else {
								b.w = uint8(min(w, b.w))
							}
						}
						ret = append(ret, b)
					}
				}
			}
		}
	}
	return ret
}

// calculates box score on pallet, which is the percentage of box circumference
// touching other boxes or pallet
func boxScoreOnPallet(b box, pg palletgrid) float64 {
	if b.w == 0 || b.l == 0 {
		return 0
	}
	circumference := (b.w + b.l) * 2
	touching := uint8(0)

	if b.x == 0 {
		touching += b.w
	} else if b.x > 0 {
		for j := b.y; j < b.y+b.w; j++ {
			if pg[(b.x-1)*palletLength+j] != emptybox {
				touching++
			}
		}
	}

	if b.x+b.l == palletWidth {
		touching += b.w
	} else if b.x+b.l < palletWidth {
		for j := b.y; j < b.y+b.w; j++ {
			if pg[(b.x+b.l)*palletLength+j] != emptybox {
				touching++
			}
		}
	}

	if b.y == 0 {
		touching += b.l
	} else if b.y > 0 {
		for i := b.x; i < b.x+b.l; i++ {
			if pg[i*palletLength+b.y-1] != emptybox {
				touching++
			}
		}
	}

	if b.y+b.w == palletLength {
		touching += b.l
	} else if b.y+b.w < palletLength {
		for i := b.x; i < b.x+b.l; i++ {
			if pg[i*palletLength+b.y+b.w] != emptybox {
				touching++
			}
		}
	}
	return (float64(touching) / float64(circumference)) * 100.0
}

func newRepacker(in <-chan *truck, out chan<- *truck) *repacker {
	r := &repacker{}
	go func() {
		boxesToPack := []box{}
		for t := range in {
			//fmt.Printf("Truck: %d\n", t.id)
			for _, p := range t.pallets {
				boxesToPack = append(boxesToPack, p.boxes...)
			}
			//fmt.Printf("Boxes to pack: %d\n", len(boxesToPack))
			packedPallets := packBoxesToPallets(boxesToPack)
			if t.id == idLastTruck {
				// last truck, load all pallets
				t.pallets = packedPallets
			} else {
				//pack only full pallets on truck, remainig boxes try to pack on next truck
				t.pallets = []pallet{}
				boxesToPack = []box{}
				remainingPallets := []pallet{}
				for _, p := range packedPallets {
					if getFreeSpace(p) == 0 {
						t.pallets = append(t.pallets, p)
					} else {
						remainingPallets = append(remainingPallets, p)
						boxesToPack = append(boxesToPack, p.boxes...)
					}
				}

				// if there are too many left boxes, pack pallets with lowest freespace on truck too
				// this is cause the algorithm gets slower with more boxes
				// this should not be, cause the algorithm should pack full pallets,
				// but just in case, the pallets have odd dimensions and boxes have even dimensions
				// the value 500 has been choosen experimentaly by benchmarks
				// TODO maybe do it other way, save packing times and allow only short packing times
				// PS: this is a ugly solution... we're loosing profit
				for len(boxesToPack) > 500 {
					//fmt.Println("More than 500 boxes remaining")
					//get smallest free space
					minFreeSpace := emptyPalletFreeSpace
					for _, p := range remainingPallets {
						freeSpace := getFreeSpace(p)
						if minFreeSpace > freeSpace {
							minFreeSpace = freeSpace
						}
					}
					//fmt.Println("minFreeSpace: ", minFreeSpace)

					remainingPallets1 := []pallet{}
					boxesToPack = []box{}
					for _, p := range remainingPallets {
						if getFreeSpace(p) == minFreeSpace {
							t.pallets = append(t.pallets, p)
						} else {
							remainingPallets1 = append(remainingPallets1, p)
							boxesToPack = append(boxesToPack, p.boxes...)
						}
					}
					remainingPallets = remainingPallets1
				}
			}
			//fmt.Printf("Out truck: %v\n", t)
			//for pn, p := range t.pallets {
			//  fmt.Printf("pallet: %d\n%s\n\n------------\n\n", pn, p)
			//}
			out <- t
		}
		// The repacker must close channel out after it detects that
		// channel in is closed so that the driver program will finish
		// and print the stats.
		close(out)
	}()
	return r
}
