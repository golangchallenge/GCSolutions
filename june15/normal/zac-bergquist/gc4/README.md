Go Challenge 4
==============

I abstracted repacker into an interface and provided several implementations.

These implementations were inspired by Jukka Jylanki's "A Thousand ways to Pack
the Bin - A Practical Approach to Two-Dimensional Rectangle Bin Packing."

See http://clb.demon.fi/files/RectangleBinPack.pdf

## Algorithms

### Shelf Repacker

This repacker uses a simple shelf first-fit algorithm.  This algorithm is the
simplest variant of the shelf algorithms.  It performs well and runs quite fast,
but only utilizes ~60% of the total pallet space.  The net profits were actually
negative because this repacker required more pallets than the input.

### Guillotine Repacker

The guillotine algorithm divides the free space in the pallet up into a set
of "free rectangles" where a box can be placed.  When a box is placed inside
a free rectangle, the remaining space in the free rectangle is split into 2 free
rectangles.

Our implementation chooses to place the box in the free rectangle with the smallest
area.  We also implement the "rectangle merge" improvement.  Each time a box is
added, we check to see if any of the free rectangles can be merged into a single
(larger) free rectangle.  This is advantageous because the guillotine algorithms
always place boxes inside a free rectangle - a box never straddles multiple free
rectangles.  

### Multi Guillotine Repacker

The standard guillotine repacker only works on a single pallet at a time.  
If a box doesn't fit on the pallet, it adds the pallet to the truck and starts
a new pallet.

The multi guillotine repacker aggregates several guilotine repackers.  When adding
a box, it checks each repacker in the set.  It only adds the pallet to a truck and
starts a new truck when there isn't room on any of the repackers.  The goal is to
avoid starting a new pallet right away on the chance that we can pack another box
or two onto the current pallet.

#### Potential Improvement

Eventually each repacker in the set gets sufficiently full such that the probability
of a box fitting in it is quite low, and the time spent checking each pallet in the
set is wasted.  I implemented a check to "close" a pallet when its free area falls
below some configurable threshold, hoping to reduce the amount of time spent checking
mostly full pallets.

This change didn't demonstrate any significant improvement.  I suspect that the
savings would be greater with larger pallets.  At the current 4x4 pallet size,
even when there's a good portion of free space, many boxes aren't going to fit.


## Tests

Each repacker comes with a set of tests.

Simply run `go test` in the `gc4` directory to run them.

## Driver Modifications

I did modify main.go and pallet.go even though the challenge prohibits this.
The modifications add an additional metric to the `result` struct -
`palletUtilization`.  

This was helpful in order to swap in various `repacker` implementations and
see which make better usage of the pallets.

These modifications don't provide any competitive advantage, and the application
will still compile if they are replaced with the original files.
