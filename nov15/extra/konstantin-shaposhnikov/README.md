## Usage

In addition to printing the solution to the standard output the application has
a terminal based UI that shows process of solving a puzzle step by step. Use
`-ui` flag to enable it.

### Demo 1: unsolveable puzzle

![unsolveable puzzle](sample.gif)

### Demo 2: level 7 puzzle

![level 7 puzzle](hard.gif)

## Implementation details

Instead of implementing brute force guessing algorithm the application tries to
apply set of rules to remove impossible candidates for each cell of a grid
(similar to a human).

The difficulty of a puzzle (level 1 to 7) is determined by the most advanced
rule that has been used to solve it:

* a puzzle of level 1 requires only a single possibility rule
* a puzzle of level 7 requires X-Wing

## Limitations

* There are known rules exist that have't been implemented (e.g. naked triples,
  hidden triples, swordfish, xy-wing, etc).
* While possible to add guessing has not implemnted.
* Unfortunately the example from the problem definition cannot be solved by the
  application (I wonder if it can be solved without guessing at all).
