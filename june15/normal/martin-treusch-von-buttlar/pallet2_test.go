package main

import "testing"

var (
	emptyPallet = pallet{boxes: []box{}}
	fullPallet  = palletMustCompileFromString("0 0 4 1 0,1 0 4 1 0,2 0 4 1 0,3 0 4 1 0")
	maxSquare   = palletLength * palletLength
)

func palletMustCompileFromString(def string) pallet {
	p, err := palletFromString(def)
	if err != nil {
		panic(err)
	}
	return p
}

func TestEmptyBoxes(t *testing.T) {
	if emptyPallet.IsFull() || emptyPallet.Items() > 0 {
		t.Error("expected empty pallet")
	}
	for w := 1; w <= palletWidth; w++ {
		for l := 1; l <= palletLength; l++ {
			p := pallet{
				boxes: []box{},
			}
			b := box{w: uint8(w), l: uint8(l), x: 0, y: 0}
			fitted := p.FitBox(b)
			if !fitted {
				t.Errorf("expected fit for: %s into\n%s\n", b, p)
			}
		}
	}
}

func TestFullBoxes(t *testing.T) {
	if !fullPallet.IsFull() {
		t.Error("expected full pallet")
	}

	for w := 1; w <= palletWidth; w++ {
		for l := 1; l <= palletLength; l++ {
			b := box{w: uint8(w), l: uint8(l), x: 0, y: 0}
			fitted := fullPallet.FitBox(b)
			if fitted {
				t.Errorf("expected no fit for: %s into\n%s\n", b, fullPallet)
			}
		}
	}
}

func TestHalfFullBoxes(t *testing.T) {
	for w := 1; w <= palletWidth; w++ {
		for l := 1; l <= palletLength; l++ {
			phor := palletMustCompileFromString("0 0 4 1 0,1 0 4 1 0")
			pver := palletMustCompileFromString("0 0 1 4 0,0 1 1 4 0")
			pouthor := palletMustCompileFromString("0 0 4 1 0,3 0 4 1 0")
			poutver := palletMustCompileFromString("0 0 1 4 0,0 3 1 4 0")

			b := box{w: uint8(w), l: uint8(l), x: 0, y: 0}
			for _, p := range []pallet{phor, pver, pouthor, poutver} {
				fitted := p.FitBox(b)
				area := int(b.l) * int(b.w)
				if area <= maxSquare/2 {
					if !fitted {
						t.Errorf("expected fit for: %s into\n%s\n", b, p)
					}
				} else if area > maxSquare/2 {
					if fitted {
						t.Errorf("expected no fit for: %s into\n%s\n", b, p)
					}
				}
			}
		}
	}
}

func BenchmarkFitBox(b *testing.B) {
	phor := palletMustCompileFromString("0 0 4 1 0,1 0 4 1 0")
	pver := palletMustCompileFromString("0 0 1 4 0,0 1 1 4 0")
	pouthor := palletMustCompileFromString("0 0 4 1 0,3 0 4 1 0")
	poutver := palletMustCompileFromString("0 0 1 4 0,0 3 1 4 0")
	poutter := palletMustCompileFromString("0 0 4 1 0,1 0 1 2 0,1 3 1 2 0,3 0 4 1 0")

	pallets := []pallet{phor, pver, pouthor, poutver, poutter}
	for n := 0; n < b.N; n++ {
		box := genbox()
		for _, p := range pallets {
			p.FitBox(box)
		}
	}
}

func BenchmarkFillGrid(b *testing.B) {
	phor := palletMustCompileFromString("0 0 4 1 0,1 0 4 1 0")
	pver := palletMustCompileFromString("0 0 1 4 0,0 1 1 4 0")
	pouthor := palletMustCompileFromString("0 0 4 1 0,3 0 4 1 0")
	poutver := palletMustCompileFromString("0 0 1 4 0,0 3 1 4 0")
	poutter := palletMustCompileFromString("0 0 4 1 0,1 0 1 2 0,1 3 1 2 0,3 0 4 1 0")

	pallets := []pallet{phor, pver, pouthor, poutver, poutter}
	for n := 0; n < b.N; n++ {
		for _, p := range pallets {
			p.fillGrid()
		}
	}
}
