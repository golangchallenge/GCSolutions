package main

import "fmt"

//Tree is a simple binary tree data structure that subdivides the pallet's remaining space based on the best fit of the box, including rotations. The root tree's rectangle represents a single pallet. All child nodes are either a subdivision of the pallet's space or a box that takes up the pallet's space.
type Tree struct {
	R      Rectangle
	lnode  *Tree
	rnode  *Tree
	hasVal bool   //space represented by R is taken up by a box
	pal    pallet //The pallet this tree represents. Only the root needs one
}

//NewTree returns the root node. This is the way trees are used from outside this file
func NewTree() *Tree {
	pal := pallet{boxes: make([]box, 0)}
	return &Tree{R: Rectangle{0, 0, palletWidth, palletLength}, hasVal: false, pal: pal}
}

func emptyTree() *Tree {
	return &Tree{hasVal: false, R: Rectangle{}}
}

//Insert takes a slice of *Tree and a box to insert. It then iterates through each tree and attempts an insert. If no trees can fit the box, it allocates a new *Tree. That is why it must return itself.
func Insert(trees []*Tree, b *box) []*Tree {
	for _, t := range trees {
		if t.Insert(b) != nil {
			t.pal.boxes = append(t.pal.boxes, *b)
			return trees
		}
	}
	tree := NewTree()
	trees = append(trees, tree)
	n := tree.Insert(b)
	if n != nil {
		tree.pal.boxes = append(tree.pal.boxes, *b)
	}
	return trees
}

//Insert is the meat of the algorithm. It recursively descends through each Node until it reaches a leaf. Then, it either subdivides the space, creating two children in the case that the box is not a perfect fit. If the box fits perfectly, the node's rectangle becomes the box's rectangle and then the node will have no more children.
func (t *Tree) Insert(b *box) *Tree {
	if !t.IsLeaf() {
		n := t.lnode.Insert(b)
		if n != nil {
			return n
		}

		return t.rnode.Insert(b)
	}

	if t.HasVal() {
		return nil
	}
	fits, fitsr := t.R.Fits(b), t.R.FitsR(b)
	if !fits && !fitsr {
		return nil
	}
	if t.R.PerfectFit(b) {
		t.hasVal = true
		b.y, b.x = t.R.Left, t.R.Top
		return t
	}
	if t.R.PerfectFitR(b) {
		t.hasVal = true
		b.y, b.x = t.R.Left, t.R.Top
		b.w, b.l = b.l, b.w
		return t
	}

	t.lnode = emptyTree()
	t.rnode = emptyTree()

	if !fitsr {
		splitnormal(t, b)
	} else if !fits {
		splitr(t, b)
	} else {
		splitboth(t, b)
	}

	return t.lnode.Insert(b)
}

//Splits the space either vertically or horizontally based on the best fit of the box. The best fit is determined as the least amount of space left over either vertically or horiztonally after fitting a box
func splitnormal(t *Tree, b *box) {
	dw := t.R.Dx() - b.w
	dh := t.R.Dy() - b.l

	if dw > dh {
		t.lnode.R = Rectangle{
			t.R.Left, t.R.Top,
			t.R.Left + b.w, t.R.Bottom,
		}
		t.rnode.R = Rectangle{
			t.R.Left + b.w, t.R.Top,
			t.R.Right, t.R.Bottom,
		}
	} else {
		t.lnode.R = Rectangle{
			t.R.Left, t.R.Top,
			t.R.Right, t.R.Top + b.l,
		}
		t.rnode.R = Rectangle{
			t.R.Left, t.R.Top + b.l,
			t.R.Right, t.R.Bottom,
		}
	}
}

func splitr(t *Tree, b *box) {
	rotate(b)
	splitnormal(t, b)
}

func splitboth(t *Tree, b *box) {
	dw := t.R.Dx() - b.w
	dh := t.R.Dy() - b.l

	drw := t.R.Dy() - b.w
	drh := t.R.Dx() - b.l

	max := maxuint(dw, dh, drw, drh)

	switch max {
	case dw:
		splitnormal(t, b)
	case dh:
		splitnormal(t, b)
	case drw:
		splitr(t, b)
	case drh:
		splitr(t, b)
	}
}

//IsLeaf returns true if the node is a leaf. A leaf is either space completely uninhabited by a box or space fully inhabited by a box
func (t *Tree) IsLeaf() bool {
	if t.hasVal == true || t.lnode == nil {
		return true
	}
	return false
}

//HasVal returns true if the space represented by the current node is filled by a box
func (t *Tree) HasVal() bool {
	if t.hasVal == true {
		return true
	}
	return false
}

//String pretty prints a tree. It's not really that pretty
func (t *Tree) String() string {
	return fmt.Sprintf("|%d %d %d %d\n|-l-%s\n-r-%s|\n", t.R.Left, t.R.Top, t.R.Right, t.R.Bottom, t.lnode, t.rnode)
}

func rotate(b *box) {
	b.l, b.w = b.w, b.l
}

func maxuint(i ...uint8) uint8 {
	max := uint8(0)
	for _, val := range i {
		if val > max {
			max = val
		}
	}
	return max
}
