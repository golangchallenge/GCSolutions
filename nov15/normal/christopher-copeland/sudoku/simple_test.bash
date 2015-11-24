#!/bin/bash

go build

time for p in $(find puzzles unsolvable_puzzles invalid_inputs -type f); do
  echo $p
  time ./sudoku < $p
done
