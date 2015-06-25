package main

const fullGrid = bingrid(^uint16(0))
const emptyGrid = bingrid(0)

type bingrid uint16

type pos struct {
	grid bingrid
	box  box
}

// posCache caches permutations for all possible box sizes
// box size is encoded as int in the first array
var permCache [][]pos

func init() {
	if palletLength != 4 || palletWidth != 4 {
		panic("palletGrid and palletWidth must equal 4")
	}

	permCache = make([][]pos, (palletLength+1)*(palletWidth+1))
	for l := uint8(1); l <= palletLength; l++ {
		for w := uint8(1); w <= palletWidth; w++ {
			perms := generatePerms(l, w)
			if w != l {
				perms = append(perms, generatePerms(w, l)...)
			}
			permCache[l*palletWidth+w] = perms
			/*
				fmt.Print(w, l)
				for _, p := range perms {
					fmt.Printf(" %016b", p.grid)
				}
				fmt.Println()
			*/
		}
	}
}

func generatePerms(l, w uint8) []pos {
	pc := (palletLength + 1 - l) * (palletWidth + 1 - w)
	perms := make([]pos, 0, pc)
	for x := uint8(0); x < palletLength; x++ {
		for y := uint8(0); y < palletWidth; y++ {
			box := box{x, y, w, l, 0}
			bg, err := box.fillGrid(emptyGrid, 1)
			if err == nil {
				perms = append(perms, pos{bg, box})
			}
		}
	}

	return perms
}

func (b *box) fillGrid(bg bingrid, bn int) (bingrid, error) {
	if b.x+b.l > palletWidth || b.y+b.w > palletLength {
		return bg, errEdge(bn)
	}
	for i := b.x; i < b.x+b.l; i++ {
		for j := b.y; j < b.y+b.w; j++ {
			if bg&(1<<(i*palletWidth+j)) != 0 {
				return bg, errOverlap(bn)
			}
			bg = bg | (1 << (i*palletWidth + j))
		}
	}
	return bg, nil
}

func (p *pallet) fillGrid() (bingrid, error) {
	var (
		pg  bingrid
		err error
	)
	for bn, b := range p.boxes {
		pg, err = b.fillGrid(pg, bn)
	}
	return pg, err
}

func (p *pallet) FitBoxWithGrid(b box) (bool, bingrid) {
	b = b.canon()
	pg, err := p.fillGrid()
	if err != nil {
		panic(err)
	}
	pg, fits := p.fitBox(pg, b)

	return fits, pg
}

func (p *pallet) FitBox(b box) bool {
	fits, _ := p.FitBoxWithGrid(b)
	return fits
}

func (p *pallet) IsFull() bool {
	pg, _ := p.fillGrid()
	return pg.IsFull()
}

func (bg bingrid) IsFull() bool {
	return bg&fullGrid == fullGrid
}

func (p *pallet) fitBox(pg bingrid, b box) (bingrid, bool) {
	fits := false
	var err error
	for _, perm := range permCache[b.l*palletWidth+b.w] {
		fits = pg&perm.grid == 0
		if fits {
			nb := perm.box
			nb.id = b.id
			p.boxes = append(p.boxes, nb)
			pg, err = p.fillGrid()
			if err != nil {
				panic(err)
			}
			break
		}
	}
	return pg, fits
}
