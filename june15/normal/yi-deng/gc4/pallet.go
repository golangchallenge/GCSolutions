package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

// A pallet holds a collections of boxes, each in a certain place on a grid.
type pallet struct {
	boxes []box
}

const palletWidth = 4
const palletLength = 4

// palletFromString reads a pallet from a string. A pallet is a comma-separated
// list of boxes.
func palletFromString(in string) (pallet, error) {
	p := pallet{}

	boxes := strings.Split(in, ",")
	for _, s := range boxes {
		b, err := boxFromString(s)
		if err != nil {
			return pallet{}, err
		}
		p.boxes = append(p.boxes, b)
	}
	return p, nil
}

func (p pallet) Items() int { return len(p.boxes) }

// IsValid returns nil if the pallet is correctly packed, otherwise an error
// that indicates the problem.
func (p pallet) IsValid() error {
	_, err := p.paint()
	return err
}

var emptybox = box{}

// paint iterates through all the boxes in a pallet
// and attempts to put them onto a palletgrid. If a box overlaps
// another, it continues painting and returns an error. If a
// box falls outside the pallet it is truncated and an error
// is returned.
func (p pallet) paint() (g palletgrid, err error) {
	for bn, b := range p.boxes {
		for i := b.x; i < b.x+b.l; i++ {
			for j := b.y; j < b.y+b.w; j++ {
				ok := true

				// Out of bounds?
				if i >= palletWidth || j >= palletLength {
					err = errEdge(bn)
					ok = false
					continue
				}

				// Was this spot already painted?
				if g[i*palletLength+j] != emptybox {
					err = errOverlap(bn)
					ok = false
				}
				if ok {
					g[i*palletLength+j] = b
				}
			}
		}
	}
	return
}

type errOverlap int

func (e errOverlap) Error() string {
	return fmt.Sprintf("box %v overlaps others", int(e))
}

type errEdge int

func (e errEdge) Error() string {
	return fmt.Sprintf("box %v goes off the edge", int(e))
}

var errEmpty = errors.New("empty box")
var errZeroBox = errors.New("zero-sized box")

const symbols = "!@#$%^&*-=+:<>?x"

func (p pallet) String() (out string) {
	// Pick a symbol to represent each box
	tochar := make(map[box]string)
	for i, x := range p.boxes {
		tochar[x] = string(symbols[i])
	}

	pg, _ := p.paint()
	out = "\n"
	for i := 0; i < palletWidth; i++ {
		out += "| "
		for j := 0; j < palletLength; j++ {
			b := pg[i*palletLength+j]
			if b == emptybox {
				out += "  "
			} else {
				out += fmt.Sprintf("%s ", tochar[b])
			}
		}
		out += "|\n"
	}
	return
}

// OneLine formats a pallet as one line, in the same format as the input.
func (p pallet) OneLine() string {
	out := make([]string, p.Items())
	for i, b := range p.boxes {
		out[i] = b.String()
	}
	return strings.Join(out, ",")
}

type palletgrid [16]box

// A box is a box, including its position on the pallet. Its
// ID is unique across all the boxes in one input file.
type box struct {
	x, y uint8
	w, l uint8
	id   uint32
}

func (b box) String() string {
	return fmt.Sprintf("%v %v %v %v %v", b.x, b.y, b.w, b.l, b.id)
}

// canon makes a canonicalized form of the box for use
// as the key in a map. The position is zeroed, and the orientation
// of the box is "horizontal" (i.e. width > length).
func (b box) canon() (out box) {
	out = b
	out.x, out.y = 0, 0
	if out.w < out.l {
		out.l, out.w = out.w, out.l
	}
	return
}

// boxFromString returns the box defined by a string of the form "x y w h id".
func boxFromString(in string) (b box, err error) {
	_, err = fmt.Sscanln(in, &b.x, &b.y, &b.w, &b.l, &b.id)
	if b == emptybox {
		return b, errEmpty
	}
	if b.w == 0 && b.l == 0 {
		return b, errZeroBox
	}
	return
}

// A truckReader scans an io.Reader, returning the trucks parsed from the input.
//
// A truck starts with "truck <id>", and ends with "endtruck". Inside of a truck,
// there's one pallet per line.
type truckReader struct {
	scn *bufio.Scanner
	err error
}

type truck struct {
	id      int
	pallets []pallet
}

const idLastTruck = 0

func newTruckReader(r io.Reader) *truckReader {
	return &truckReader{
		scn: bufio.NewScanner(r),
	}
}

func (r *truckReader) Next() (*truck, error) {
	if r.err != nil {
		return nil, r.err
	}

	t := &truck{}
	for {
		if r.scn.Scan() == false {
			r.err = r.scn.Err()
			if r.err == nil {
				r.err = io.EOF
			}
			return nil, r.err
		}

		if strings.HasPrefix(r.scn.Text(), "truck") {
			var truck string
			_, r.err = fmt.Sscanln(r.scn.Text(), &truck, &t.id)
			if r.err != nil {
				return nil, r.err
			}
			continue
		}

		if r.scn.Text() == "endtruck" {
			break
		}

		var p pallet
		p, r.err = palletFromString(r.scn.Text())
		if r.err != nil {
			return nil, r.err
		}
		t.pallets = append(t.pallets, p)
	}
	return t, r.err
}
