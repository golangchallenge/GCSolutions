package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

/*
 * The solution presented here relies on a minimal and adjusted 
 * version of an evolutionary algorithm (EA), combined with a function
 * to compress boxes on pallets (tetris shift).
 * The algorithm works as follows: An initial allocation of 
 * boxes to pallets is produced by simply packing each box on one
 * pallet. After that, the tetris-shift-function compresses this 
 * allocation by - vividly explained - shifting the boxes just as
 * in the eponymous game. Since for bad input data this deterministic
 * function could produce a highly fragmented solution, a 
 * non-deterministic component is added by randomly selecting a 
 * box, repacking everything except this box and re-adding the box 
 * at the end of the truck. Using this technique, four independent 
 * solutions are generated among which the best one is selected and
 * finally returned.
 * 
 * The presented solution has two weaknesses, both intentionally 
 * implemented in order to keep the code maintainable and extensible
 * as demanded by the customer:
 * 1.   The EA has no classical crossover operator like conventional 
 * EAs. Such a crossover operator is possible and would augment the
 * solution slightly, but complicate the code disproportionately. Thus,
 * the EA also only computes one follow-up generation, as more 
 * generations would make no sense without crossover.
 * 2.   The repacking takes place for each truck individually, no boxes
 * are shifted between trucks. Implementing this would require storing
 * all boxes from every truck from the input data, which would make the
 * code significantly more complex.
 * 
 * Summarizing these weaknesses, it can be stated that solutions to
 * these issues are theoretically possible though impractical for the
 * given requirements. Tests have shown that the presented solution
 * yields allocations quite near to the optimal solutions, thus, the 
 * presented solution is a reasonable tradeoff between profit and 
 * code maintainability.
 * /

/* Types and constants: */

// A repacker repacks trucks.
type repacker struct {
}

//4x4 square grid, stores box id
const gridSize = 4

//a grid of pointers to store which box occupies which position on the
//pallet
type grid2d [gridSize][gridSize]*box

//a gene represents an allocation of boxes to pallets
type gridGene []grid2d

//probability that a box is rotated during mutation
const rotationProbability = 50

//size of a generation for the EA
const generationSize = 5






/* Convenience functions to print (intermediate) results: */

// Print all pallets on the truck 
func (g gridGene) print() {
	fmt.Printf("****************\n")
	for _, g1 := range g {
		g1.print()
	}
	fmt.Printf("****************\n")
}

// Output the pallets on the truck in SVG format to produce
// a visualization
func (g gridGene) printSVG() {
	fmt.Printf("<svg  xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\">\n")

	for i, g1 := range g {
		g1.printSVG(i)
	}

	fmt.Printf("\n")
	fmt.Printf("<rect x=\"%d\" y=\"%d\" height=\"%d\" width=\"%d\" style=\"stroke:#000000; fill: #ffffff\"/>", 0, len(g)*gridSize*20, gridSize*10, 20*gridSize)
	info := "Fitness: " + strconv.Itoa(g.fitness())
	fmt.Printf("<text x=\"%d\" y=\"%d\" style=\"text-anchor: start\">%s</text>", 10, len(g)*gridSize*20+10, info)

	fmt.Printf("</svg>\n")
}

// Print one individual pallet. Numbers indicate the boxes' ids 
// on the spots they occupy, or 0 if a spot on the pallet is free
func (g grid2d) print() {
	fmt.Printf("-----\n")
	for i := 0; i < gridSize; i++ {
		for j := 0; j < gridSize; j++ {
			if g[i][j] != nil {
				fmt.Printf("%d ", g[i][j].id)
			} else {
				fmt.Printf("0 ")
			}
			if (j + 1) == gridSize {
				fmt.Printf("\n")
			}
		}
	}
	fmt.Printf("-----\n")
}

// Prints a rendering of an individual pallet in SVG format. 
// Numbers indicate the boxes' ids on the spots they occupy, 
// or 0 if a spot on the pallet is free
func (g grid2d) printSVG(offset int) {

	size := 20
	off := offset * size * gridSize

	fmt.Printf("\n")
	for i := 0; i < gridSize; i++ {
		for j := 0; j < gridSize; j++ {
			if g[i][j] != nil {
				//fmt.Printf("%d ",g[i][j].id)
				fmt.Printf("<rect x=\"%d\" y=\"%d\" height=\"%d\" width=\"%d\" style=\"stroke:#000000; fill: #c0c0c0\"/>", j*size, i*size+off, size, size)
				fmt.Printf("<text x=\"%d\" y=\"%d\" style=\"text-anchor: start\">%d</text>", j*size+(size/8), i*size+off+(size/2), g[i][j].id)
			} else {
				fmt.Printf("<rect x=\"%d\" y=\"%d\" height=\"%d\" width=\"%d\" style=\"stroke:#000000; fill: #ffffff\"/>", j*size, i*size+off, size, size)
			}
			if (j + 1) == gridSize {
				fmt.Printf("\n")
			}
		}
	}
	fmt.Printf("<line x1=\"%d\"  y1=\"%d\" x2=\"%d\" y2=\"%d\" style=\"stroke:#000000;stroke-width:3;\"/> \n", 0, off+size*gridSize, size*gridSize, off+size*gridSize)
}









/* Functions for actual optimization: */

// Compute the quality of a given solution.
// The quality computed here is a relative measure as it simply counts
// the empty pallets. To compare solutions generated in the same manner,
// this is a useful measure since more empty pallets yield more profit.
func (g gridGene) fitness() int {
	fitness := 0
	for _, g2 := range g {
		if g2.isEmpty() {
			fitness++
		}
	}
	return fitness
}

// Find a free spot to insert the respective box.
// Prefer cells which are 1. high and 2. leftmost
func (g grid2d) freeSpot(minWidth uint8, minHeight uint8, bx box) (col uint8, row uint8, rotate bool) {
	for startI := 0; startI < gridSize; startI++ {
		for startJ := 0; startJ < gridSize; startJ++ {
			//use each matrix cell as potential upper left corner of box
			//for each of these starting points, check whether sufficient
			//space is available in x- and y-direction
			
			var freeCountWithoutRotation uint8 
			var freeCountWithRotation uint8 
			
			// Check if the box would fit without rotation
			for i := startI; i < gridSize && i-startI < (int)(minHeight); i++ {
				for j := startJ; j < gridSize && j-startJ < (int)(minWidth); j++ {
					if g[i][j] == nil {
						//empty cell
						freeCountWithoutRotation++
					} 
				}
			}
			
			// Check if the rotation of the box would fit
			for i := startI; i < gridSize && i-startI < (int)(minWidth); i++ {
				for j := startJ; j < gridSize && j-startJ < (int)(minHeight); j++ {
					if g[i][j] == nil {
						//empty cell
						freeCountWithRotation++
					} 
				}
			}

			if freeCountWithoutRotation == minWidth*minHeight {
				// Fits without rotation. 
				return (uint8)(startI), (uint8)(startJ), false
			} else {
				if freeCountWithRotation == minWidth*minHeight {
					// Fits if rotated. 
					return (uint8)(startI), (uint8)(startJ), true
				}
			} 
		}
	}

	return (uint8)(gridSize + 1), (uint8)(gridSize + 1), false
}

// This function compresses the boxes on the truck. It tries to shift
// boxes from the end to the front of the truck by checking where
// the foremost position is to place the box.
func (g gridGene) tetrisShift() {
	for i := 0; i < len(g)-1; i++ {
		for j := i + 1; j < len(g)-1; j++ {
			// Try to shift from pallet j to pallet i
			g.shift(j, i)
		}
	}
	// The last grid is the extracted one from the mutation
	// (or empty for the initial iteration
	for j := 0; j < len(g)-1; j++ {
		g.shift(len(g)-1, j)
	}
}

// Function that tries to shift all boxes from pallet with id "from" to 
// pallet with id "to"
func (g gridGene) shift(from int, to int) {
	g2 := g[from]
	cont := g2.getContainedBoxes()
	for _, value := range cont {
		ffr, ffc, needsRotation := g[to].freeSpot((*value).l, (*value).w, (*value))
		if ffr <= gridSize && ffc <= gridSize {
			// If first free row/column is inside of the bounds given
			// by gridSize, the box can be inserted at the given position
			if(!needsRotation) {
				g[to].addBox(*value, ffc, ffr)
			} else {
				// If the box only fits after a rotation, width and 
				// length have to be switched
				(*value).w,(*value).l = (*value).l,(*value).w
				g[to].addBox(*value, ffc, ffr)
			}
			g[from].removeBox(value)
		} 
	}
}

// Returns a new allocation of boxes to pallets by copying the 
// original allocation and carrying out a mutation operation
func (g *gridGene) getMutation() (gnew gridGene) {
	gg := g.getDeepCopy()
	gg.mutation()
	return gg
}

// Computes a mutation of the given allocation by randomly chosing
// a box which is re-allocated
func (g *gridGene) mutation() {
	rand.Seed(time.Now().UTC().UnixNano())
	var idfrom int
	for {
		idfrom = rand.Intn(len(*g))
		if !(*g)[idfrom].isEmpty() {
			break
		}
	}
	gfrom := &(*g)[idfrom] // pallet from which to choose the box
	
	var b *box
	for {
		i := rand.Intn(gridSize)
		j := rand.Intn(gridSize)
		b = gfrom[i][j] //large boxes have higher probability of getting re-allocated
		if b != nil {
			break
		}
	}

	rotation := rand.Intn(100)
	if rotation < rotationProbability {
		// randomly rotate some boxes
		b.w, b.l = b.l, b.w
	}

	// shift box to the new position
	(*g)[(len(*g)-1)].addBox(*b, 0, 0)
	gfrom.removeBox(b)
}








/* Helper functions to work with pallets or allocations of boxes: */

// Creates a deep copy of the grid gene
func (g *gridGene) getDeepCopy() (gnew gridGene) {
	gg := make(gridGene, len(*g), len(*g))
	for i, g2d := range *g {
		cont := g2d.getContainedBoxes()
		for _, value := range cont {
			// Create a completely new object
			b := box{id: (*value).id, x: (*value).x, y: (*value).y, l: (*value).l, w: (*value).w}
			gg[i].addBox(b, (b).x, (b).y)
		}
	}
	return gg
}


// Checks wether a pallet is completely empty
func (g grid2d) isEmpty() bool {
	for i := 0; i < gridSize; i++ {
		for j := 0; j < gridSize; j++ {
			if g[i][j] != nil {
				// Check wether each cell only contains the nil-pointer
				return false
			}
		}
	}
	return true
}

// Removes the referenced box from the truck
func (g *grid2d) removeBox(b *box) {
	for i := 0; i < gridSize; i++ {
		for j := 0; j < gridSize; j++ {
			if g[i][j] == b {
				g[i][j] = nil
			}
		}
	}
}

// Adds a box at the specified position
func (g *grid2d) addBox(b box, x uint8, y uint8) {
	for i := y; i < (y + b.w); i++ {
		for j := x; j < (x + b.l); j++ {
			if i+1 > gridSize || j+1 > gridSize {
				return // error handling
			}
			g[i][j] = &b
			g[i][j].x = x
			g[i][j].y = y
		}
	}
}

// Return all boxes contained on the truck
func (g *grid2d) getContainedBoxes() (contained map[uint32]*box) {
	m := make(map[uint32]*box)
	for i := 0; i < gridSize; i++ {
		for j := 0; j < gridSize; j++ {
			if g[i][j] == nil {
				continue
			}
			_, ok := m[g[i][j].id]
			if !ok {
				m[g[i][j].id] = g[i][j]
			}
		}
	}
	return m
}

// The immproved version of repacking
func betterRepacking(t *truck) (out *truck) {
	fmt.Printf("Repacking truck nr %d \n", t.id)
	
	out = &truck{id: t.id}
	genes := [generationSize]gridGene{} // the concurrently existing allocations 
	boxes := []box{}

	// Build the initial allocation: each box on a distinct pallet
	for _, p := range t.pallets {
		for _, b := range p.boxes {
			g2 := grid2d{}
			g2.addBox(b, (0), (0))
			genes[0] = append(genes[0], g2)
		}
	}

	if len(genes[0]) > 0 {
		genes[0] = append(genes[0], grid2d{}) //adding an extra empty pallet for the mutation operator
		for k := 1; k < len(genes); k++ {
			genes[k] = append(genes[k], grid2d{}) //adding an extra empty grid for the mutation operator
		}
		// Compress the allocation on the initial allocation
		genes[0].tetrisShift()

		// For all other allocations, carry out a mutation and compress again
		for k := 1; k < len(genes); k++ {
			genes[k] = genes[0].getMutation()
			genes[k].tetrisShift()
		}

		// Compute the quality of the solutions and choose the best one
		// for the result
		maxFitness := genes[0].fitness()
		maxFitnessGene := 0
		for k := 0; k < len(genes); k++ {
			fitness := genes[k].fitness()
			if fitness > maxFitness {
				maxFitness = fitness
				maxFitnessGene = k
			}
		}
		
		// Return the best solution
		for _, g2 := range genes[maxFitnessGene] {

			if g2.isEmpty() {
				continue
			}
			cont := g2.getContainedBoxes()
			for _, value := range cont {
				boxes = append(boxes, (*value))
			}
			out.pallets = append(out.pallets, pallet{boxes})
			boxes = []box{}
		}
	}

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

			//			t = oneBoxPerPallet(t)
			t = betterRepacking(t)
			out <- t
		}
		// The repacker must close channel out after it detects that
		// channel in is closed so that the driver program will finish
		// and print the stats.
		close(out)
	}()
	return &repacker{}
}

// This repacker is the worst possible, since it uses a new pallet for
// every box. Your job is to replace it with something better.
func oneBoxPerPallet(t *truck) (out *truck) {
	out = &truck{id: t.id}
	for _, p := range t.pallets {
		for _, b := range p.boxes {
			fmt.Printf("% d %d \n", b.x, b.y)

			b.x, b.y = 0, 0
			out.pallets = append(out.pallets, pallet{boxes: []box{b}})
		}
	}
	return
}
