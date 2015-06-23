Introduction
============
In the logistics industry poorly packed boxes are wasting space, and worse,
tying up pallets that could be resold for a profit. We are in charge of
programming robots to repack the boxes.

You are given a truck full of a pallets with boxes on them that may or may
not be correctly packed. Your task is to implement an algorithm that packs
boxes onto the pallets correctly. A correctly packed pallet is one where
none of the boxes on it overlap, and none of the boxes hang over the edge of
the pallet. Pallets are packed in only two dimensions, with a single layer
of boxes which have an arbitrary height. All of the trucks are going the
same place anyway, so it doesn't matter which truck a box goes in, as long
as it is packed correctly on a pallet.

Empty pallets left over after repacking are pure profit. More empties = more
better! And if a truck leaves the warehouse with more pallets on it than it
came with, it comes out of your profit. So pack carefully!


Boxes
=====

A box is a `box struct`, including its position on the pallet `x`,`y` and
its width and length `w`,`l`. Its `id` is unique across all the boxes in one
input file.
```
type box struct {
	x, y uint8
	w, l uint8
	id   uint32
}
```

In their canonical form `b.canon()` a box is horizontal `w>h`:

```
  +--------+
h |        |
  +--------+
      w
```
Possible box types with respected size = l * w (using hex values `c=12` and
`f=16`) are: 
```
1 22 333 4444

44 666 8888
44 666 8888

999 cccc
999 cccc
999 cccc

ffff
ffff
ffff
ffff
```
These are all boxes that can be placed on a 4x4 pallet. However, the input
will give us boxes that are even bigger than `f`. These have to be filtered
out.

Storing boxes
-------------
Note, that an area uniquely identifies the box type, except for an area of
4. This suggests, we can use the box size as a hash and store the boxes in a
hash table. To handle the 'collision' of size 4, we can use the hash `4` for
4x1 and the hash `5` for 2x2 boxes.

![hash tab](figures/hashtab.png)

For each hash value we will have a list of boxes. If a box is repacked on
pallet, it gets pulled from the list. If a new truck comes, the new boxes
will be added.

**TODO**: When unloading a truck full of boxes it might be faster to first
store the boxes in an array and then sort the array after the size of the
box. Then transform that sorted array into a hash table.

### The Box list
The box list will operate as a last-in-first-out (LIFO) stack. Operations we
need are 
- `newItem` to create a new item
- `push` to add a new item to the front 
- `pop` to delete an item from the front

Palettes
========
A pallet holds a collections of `boxes`, each in a certain place on a grid.
```
type pallet struct {
	boxes []box
}
```
All pallets have
```
const palletWidth = 4
const palletLength = 4
```
A palette string is a comma separated list of boxes and look like this:
```
0 0 1 1 101, 1 1 1 1 102, 2 2 1 1 103, 3 0 4 1 104
```
This particular pallet could be visualized as follows:
```
| @       |
|   &     |
|     #   |
| $ $ $ $ |
```

Trucks
======

A truck has an unique `id` and contains a slice of `pallets`.
```
type truck struct {
	id      int
	pallets []pallet
}
```
A truck `string` starts with `truck <id>`, and ends with `endtruck`. Inside
of a truck, there's one pallet per line.
```
truck 1
0 0 1 1 101,1 1 1 1 102,2 2 1 1 103,3 0 4 1 104
0 0 1 1 101,0 0 1 1 102
0 0 5 5 101
endtruck
```

Functions
=========

Function `paint` will take a pallet (as a list of boxes) and tries to fill a
pallet grid with them.

Repackaging trucks
==================

Pack pallets as tight as possible. If a pallet is not full, hold it back
until it can be filled nicely and put it on the next truck.

1. Truck comes in
2. Unload truck and create boxes hashtable, see [above](#storing-boxes)
3. Run repackaging algorithm

Algorithm Idea
--------------

- We start with an empty 4x4 pallet as a free grid to place a box on. 

- Each new box is placed into the lower left of a grid.

- This divides the remaining free space into two new grids: one above and one the right.

- Again, we then try to place the next box into the most left of the two grids. 

- When a grid is completely filled its left child should give a `nil`. We
	then back up one node and proceed the right child.

  ![Free space tree structure](figures/tree.png)

  When we joint the open ends of the tree we get the total remaining free space.

  ![Combined free space](figures/add-free-space.png)

### Grid
This suggest as a grid data element the following:
```
type gridElement {
	x,y         int  //origin
	w,l         int  //width length
	size        int
	orientation enum //horizontal, vertical, square
}
```
And the grid itself is a tree.
``` 
type grid {
	parent      *gridElement
	left        *gridElement
	right       *gridElement
}
```

### Picking a box

- if the new grid has, e.g., a size of 8, we look at the box hash table with
	hash `8`. If the box list at `8` is empty we look downwards until we find
	a highest hash with a non-emptybox list and start to fill.

### Conflicts

However conflicts will occur, when overlapping grids are filled:

![Conflict](figures/conflict.png)

In the above example we did back up from the left most child to its parent
and then choose to fill the right grid with a 2x1 box,
yielding a `nil`, too. However filling the red area creates a conflict
with two areas further above the tree.

### Possible Solutions
- Not backing up again. Just fill grid to the most left. Leave the rest free.
- Propagate overlapping grid vertices to children.
  * The ones we back up the tree, we check, if an overlap is present. When we fill the grid, we update the `node`, which is further up the tree, with its new `x`, `y` and `w`, `l`. 
  * Or we keep an conflict switch, if it is activated keep backing further up, until we are on top of the conflict. 
```
type overlap struct {
  x,y int       // origin of overlap
  w,l int       // width, length of overlap region 
  node *Element // Element within which the overlap region resides
}
```
- Use a three-way tree to separete each grid into 3 non-overlapping regions. The upper (green), right (blue) and the former overlap (red).

  ![Three-way tree](figures/3tree.png)

  But now very small grids might occur. In the above example, if we have one more 2x1 box it could not be placed, except for we combine the grids.

### Other ideas
Although the graphics suggests a tree as a data structure, it would be more
efficient to use an ordered List, which contains only the leafs (nodes
without children) of the above tree, as these are the free grids we can fill
with boxes. 

The list will contain the free grids in decreasing order with the most left
first. We always try to fill the most left. When an element of the list
gets a box, it is pulled out of the list, and two other elements with the
remaining space are sorted into the list at the appropriate place. 

Optimizations
-------------

- If there are many boxes of the same type in the hashtable we might just
	fill up a pallet with same sized boxes:

```
if number box.size(8)%2 == 0 {
fill as many pallets with 4x2 boxes
}
```

```
if number box.size(4)%4 == 0 {
fill as many pallets with 2x2 boxes
}
```

Resources
=========
- http://golang-challenge.com/go-challenge4/
- http://0xax.blogspot.de/2014/08/binary-tree-and-some-generic-tricks.html
- http://nathanleclaire.com/blog/2014/07/19/demystifying-golangs-io-dot-reader-and-io-dot-writer-interfaces/
- https://golang.org/pkg/fmt/

