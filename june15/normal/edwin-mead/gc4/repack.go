package main

func newRepacker(in <-chan *truck, out chan<- *truck) {
	go func() {
		for t := range in {
			// repack each truck separately
			boxList := make(map[uint8][]uint32)
			unpack(t, boxList)
			out <- repack(t, boxList)
		}
		close(out)
	}()
}

// unpack all boxes from pallets on truck
func unpack(t *truck, boxList map[uint8][]uint32) {
	for _, p := range t.pallets {
		for _, b := range p.boxes {
			var lw uint8
			if b.w > b.l {
				lw = b.w*10 + b.l
			} else {
				lw = b.l*10 + b.w
			}
			// Save list of boxes, indexed by LxW
			boxList[lw] = append(boxList[lw], b.id)
		}
	}
}

// repack boxes onto pallets on truck
func repack(t *truck, boxList map[uint8][]uint32) (out *truck) {
	out = &truck{id: t.id}
	for {
		var p pallet
		switch {
		case len(boxList[44]) > 0:
			packBox(0, 0, 4, 4, &p, boxList)
		case len(boxList[43]) > 0:
			packBox(0, 0, 3, 4, &p, boxList)
			oneColumn(3, &p, boxList)
		case len(boxList[42]) > 0:
			packBox(0, 0, 2, 4, &p, boxList)
			twoColumns(2, &p, boxList)
		case len(boxList[41]) > 0:
			packBox(0, 0, 1, 4, &p, boxList)
			oneColumn(1, &p, boxList)
			twoColumns(2, &p, boxList)
		case len(boxList[33]) > 0:
			packBox(0, 0, 3, 3, &p, boxList)
			if len(boxList[31]) > 0 {
				packBox(3, 0, 3, 1, &p, boxList)
			} else if len(boxList[21]) > 0 {
				packBox(3, 0, 2, 1, &p, boxList)
				if len(boxList[11]) > 0 {
					packBox(3, 2, 1, 1, &p, boxList)
				}
			} else if len(boxList[11]) > 0 {
				packBox(3, 0, 1, 1, &p, boxList)
				if len(boxList[11]) > 0 {
					packBox(3, 1, 1, 1, &p, boxList)
					if len(boxList[11]) > 0 {
						packBox(3, 2, 1, 1, &p, boxList)
					}
				}
			}
			oneColumn(3, &p, boxList)
		case len(boxList[32]) > 0:
			packBox(0, 0, 2, 3, &p, boxList)
			if len(boxList[21]) > 0 {
				packBox(3, 0, 2, 1, &p, boxList)
			} else if len(boxList[11]) > 0 {
				packBox(3, 0, 1, 1, &p, boxList)
				if len(boxList[11]) > 0 {
					packBox(3, 1, 1, 1, &p, boxList)
				}
			}
			twoColumns(2, &p, boxList)
		case len(boxList[31]) > 0:
			packBox(0, 0, 1, 3, &p, boxList)
			if len(boxList[11]) > 0 {
				packBox(3, 0, 1, 1, &p, boxList)
			}
			oneColumn(1, &p, boxList)
			twoColumns(2, &p, boxList)
		default:
			twoColumns(0, &p, boxList)
			twoColumns(2, &p, boxList)
		}
		out.pallets = append(out.pallets, p)
		if len(boxList) == 0 {
			break
		}
	}
	return
}

// pack box onto pallet and delete from list
func packBox(x uint8, y uint8, w uint8, l uint8, p *pallet, m map[uint8][]uint32) {
	var lw uint8
	if l > w {
		lw = l*10 + w
	} else {
		lw = w*10 + l
	}
	p.boxes = append(p.boxes, box{x, y, w, l, m[lw][0]})

	if len(m[lw]) > 1 {
		m[lw] = append(m[lw][:0], m[lw][1:]...)
	} else {
		delete(m, lw)
	}
}

// oneColumn packs boxes onto specified pallet column
func oneColumn(y uint8, p *pallet, m map[uint8][]uint32) {
	switch {
	case len(m[41]) > 0:
		packBox(0, y, 1, 4, p, m)
	case len(m[31]) > 0:
		packBox(0, y, 1, 3, p, m)
		if len(m[11]) > 0 {
			packBox(3, y, 1, 1, p, m)
		}
	case len(m[21]) > 0:
		packBox(0, y, 1, 2, p, m)
		if len(m[21]) > 0 {
			packBox(2, y, 1, 2, p, m)
		} else if len(m[11]) > 0 {
			packBox(2, y, 1, 1, p, m)
			if len(m[11]) > 0 {
				packBox(3, y, 1, 1, p, m)
			}
		}
	case len(m[11]) > 0:
		for i := 0; i <= 3 && i < len(m[11]); i++ {
			packBox(uint8(i), y, 1, 1, p, m)
		}
	}
}

// twoColumns packs boxes onto columns specified by top-left co-ordinate
func twoColumns(y uint8, p *pallet, m map[uint8][]uint32) {
	switch {
	case len(m[42]) > 0:
		packBox(0, y, 2, 4, p, m)
	case len(m[41]) > 0:
		packBox(0, y, 1, 4, p, m)
		oneColumn(y+1, p, m)
	case len(m[32]) > 0:
		packBox(0, y, 2, 3, p, m)
		if len(m[21]) > 0 {
			packBox(3, y, 2, 1, p, m)
		} else if len(m[11]) > 0 {
			packBox(3, y, 1, 1, p, m)
			if len(m[11]) > 0 {
				packBox(3, y+1, 1, 1, p, m)
			}
		}
	case len(m[31]) > 0:
		packBox(0, y, 1, 3, p, m)
		if len(m[11]) > 0 {
			packBox(3, y, 1, 1, p, m)
		}
		oneColumn(y+1, p, m)
	default:
		square2x2(0, y, p, m)
		square2x2(2, y, p, m)
	}
}

// square2x2 packs boxes in a 2x2 shape onto specified co-ordinates
func square2x2(x uint8, y uint8, p *pallet, m map[uint8][]uint32) {
	switch {
	case len(m[22]) > 0:
		packBox(x, y, 2, 2, p, m)
	case len(m[21]) > 0:
		packBox(x, y, 1, 2, p, m)
		if len(m[21]) > 0 {
			packBox(x, y+1, 1, 2, p, m)
		} else if len(m[11]) > 0 {
			packBox(x, y+1, 1, 1, p, m)
			if len(m[11]) > 0 {
				packBox(x+1, y+1, 1, 1, p, m)
			}
		}
	case len(m[11]) > 0:
		packBox(x, y, 1, 1, p, m)
		if len(m[11]) > 0 {
			packBox(x, y+1, 1, 1, p, m)
			if len(m[11]) > 0 {
				packBox(x+1, y+1, 1, 1, p, m)
				if len(m[11]) > 0 {
					packBox(x+1, y, 1, 1, p, m)
				}
			}
		}
	}
}
