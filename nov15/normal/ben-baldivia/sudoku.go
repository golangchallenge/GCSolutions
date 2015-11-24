package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	SOLVED  = errors.New("Solved")
	REVERT  = errors.New("Revert")
	INVALID = errors.New("Invalid")
)

type Root struct {
	C      map[int]*Head
	CR     *Head
	Solved map[int]int
}

type Head struct {
	N int
	S int
	O *Dataobj
}

type Dataobj struct {
	C *Head
	U *Dataobj
	D *Dataobj
	L *Dataobj
	R *Dataobj
}

func main() {
	r := newRoot()
	r.getBoard()
	s := time.Now()
	r.initBoard()
	log.Println("Initialization: ", time.Since(s))
	p := time.Now()
	err := r.Dance()
	if err != nil {
		fmt.Println(err)
	}
	log.Println("Processing Time: ", time.Since(p))
}

func newNode() *Dataobj {
	n := &Dataobj{}
	n.U = n
	n.D = n
	n.L = n
	n.R = n
	return n
}

func newRoot() *Root {
	return &Root{
		C:      make(map[int]*Head),
		CR:     &Head{},
		Solved: make(map[int]int),
	}

}

func insertIntoRow(r *Dataobj, i *Dataobj) {
	i.R = r
	i.L = r.L
	r.L.R = i
	r.L = i
}

func (r *Root) insertIntoCol(i *Dataobj, idx int) {
	i.C = r.C[idx]
	i.D = r.C[idx].O
	i.U = r.C[idx].O.U
	r.C[idx].O.U.D = i
	r.C[idx].O.U = i
	r.C[idx].S++
}

func (r *Root) initBoard() {
	if len(r.Solved) < 17 {
		log.Fatalln("Insufficient hints for unique solution")
	}
	// Create constraint headers (columns)
	r.CR = &Head{
		N: 1000,
		S: 1,
		O: newNode(),
	}
	r.CR.O.C = r.CR
	for i := 0; i < (9 * 9 * 4); i++ {
		r.C[i] = &Head{
			N: i,
			S: 0,
			O: newNode(),
		}
		insertIntoRow(r.CR.O, r.C[i].O)
		r.C[i].O.C = r.C[i]
	}

	// Create rows
	// y position
	var svArr []*Dataobj
	for i := 0; i < 9; i++ {
		// x position
		for j := 0; j < 9; j++ {
			// If solution exists
			sIdx := (i * 9) + j
			v, s := r.Solved[sIdx]

			// value
			for k := 0; k < 9; k++ {
				// Create new row
				n := newNode()

				if s {
					if k == v {
						svArr = append(svArr, n)
					}
				}

				// First constraint: Col has value
				idx := (81 * 0) + (i * 9) + k
				r.insertIntoCol(n, idx)
				// Create 2 more nodes
				for ii := 0; ii < 3; ii++ {
					l := newNode()
					insertIntoRow(n, l)
				}

				// Second constraint: Row has value
				idx = (81 * 1) + (j * 9) + k
				r.insertIntoCol(n.R, idx)

				// Third constraint: Block has value
				idx = (81 * 2) + ((((i / 3) * 3) + (j / 3)) * 9) + k
				r.insertIntoCol(n.R.R, idx)

				// Fourth constraint: Cell has single value
				idx = (81 * 3) + (i * 9) + j
				r.insertIntoCol(n.R.R.R, idx)
			}
		}
	}
	for i := 0; i < len(svArr); i++ {
		svArr[i].selectRow()
	}
}

func (r *Root) getBoard() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter board:\n")
	for i := 0; i < 9; i++ {
		text, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		arr := strings.Split(text, " ")
		if len(arr) != 9 {
			log.Fatal("Invalid board invalid row length")
		}
		for j := 0; j < 9; j++ {
			v, err := strconv.Atoi(strings.Split(arr[j], "\n")[0])

			if err != nil {
				if strings.Split(arr[j], "\n")[0] == "_" {
					continue
				}
				log.Fatal(err)
			}
			if 0 < v && v < 10 {
				r.Solve(i, j, v-1)
			}
		}
	}
	return
}

func (r *Root) Solve(y, x, v int) {
	idx := (y * 9) + x
	r.Solved[idx] = v
}

func (r *Root) Unsolve(y, x, v int) {
	idx := (y * 9) + x
	delete(r.Solved, idx)
}

func (d *Dataobj) selectRow() error {
	d.coverCol()
	ref := d.R
	for ref != d {
		err := ref.coverCol()
		if err != nil {
			return err
		}
		ref = ref.R
	}
	return nil
}

func (d *Dataobj) unselectRow() {
	ref := d.L
	for ref != d {
		ref.uncoverCol()
		ref = ref.L
	}
	d.uncoverCol()
}

func (d *Dataobj) uncoverCol() {
	ref := d.C.O.U
	// Uncover column header if covered
	if d.C.O.L.R != d.C.O {
		d.C.O.L.R = d.C.O
		d.C.O.R.L = d.C.O
	}
	var f bool
	for ref != d.C.O {
		if ref == d {
			f = false
		} else {
			f = true
		}
		ref.uncoverRow(f)
		if ref != d {
			d.C.S++
		}
		ref = ref.U
	}
}

func (d *Dataobj) uncoverRow(f bool) {
	ref := d.L
	for ref != d {
		ref.U.D = ref
		ref.D.U = ref
		if f {
			ref.C.S++
		}
		ref = ref.L
	}
}

func (d *Dataobj) coverCol() error {
	ref := d.C.O.D
	// Cover column header if not covered
	if d.C.O.L.R == d.C.O {
		d.C.O.L.R = d.C.O.R
		d.C.O.R.L = d.C.O.L
	}
	var i = 0
	var f bool
	for ref != d.C.O {
		if i == 20 {
			log.Fatalln("Fail")
		}
		if d == ref {
			f = false
		} else {
			f = true
		}
		err := ref.coverRow(f)
		if err != nil {
			return err
		}
		if ref != d {
			d.C.S--
			if ref.C.S == 0 {
				return INVALID
			}
		}
		ref = ref.D
		i++
	}
	return nil
}

func (d *Dataobj) coverRow(f bool) error {
	ref := d.R

	for ref != d {
		ref.U.D = ref.D
		ref.D.U = ref.U
		if f {
			ref.C.S--
			if ref.C.S == 0 {
				return INVALID
			}
		}
		ref = ref.R
	}
	return nil
}

func (d *Dataobj) decode() (y, x, v int) {
	if d.C.N >= 243 {
		y = (d.C.N - 243) / 9
		x = (d.C.N - 243) % 9
	} else if d.C.N < 161 {
		v = (d.C.N % 9)
	}

	ref := d.R
	for ref != d {
		if ref.C.N >= 243 {
			y = (ref.C.N - 243) / 9
			x = (ref.C.N - 243) % 9
		} else if ref.C.N < 161 {
			v = (ref.C.N % 9)
		}
		ref = ref.R
	}
	return
}

func (r *Root) sScan() (*Dataobj, error) {
	if r.CR.O.R == r.CR.O {
		return &Dataobj{}, SOLVED
	}
	ref := r.CR.O.R
	sel := r.CR.O.R
	for ref != r.CR.O {
		if ref.C.S < sel.C.S {
			sel = ref
		}
		ref = ref.R
	}
	return sel, nil
}

func (r *Root) setSolved() {
	//fmt.Println("Solved:")
	a := []int{}
	for k, _ := range r.Solved {
		a = append(a, k)
	}
	sort.Ints(a)
	for _, v := range a {
		if v%9 == 0 {
			fmt.Printf("\n")
		}
		r.Solved[v]++
		fmt.Printf("%d ", r.Solved[v])
	}
	fmt.Printf("\n")
}

func (r *Root) Dance() error {
	if r.CR.O.R == r.CR.O {
		r.setSolved()
		return nil
	}
	sel, err := r.sScan()
	if err != nil {
		return err
	}
	ref := sel.D
	for ref != sel {
		r.Solve(ref.decode())
		err = ref.selectRow()
		if err != nil {
			ref.unselectRow()
			r.Unsolve(ref.decode())
			ref = ref.D
			continue
		}
		err = r.Dance()
		if err == nil {
			return nil
		}
		ref.unselectRow()
		r.Unsolve(ref.decode())
		ref = ref.D
	}
	return INVALID
}
