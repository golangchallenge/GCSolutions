package main

import (
	"sort"
	"strings"
)

func (b *board) SolveSimpleColoring() error {
	const technique = "SIMPLE-COLORING"

	var err error

valueLoop:
	for v := uint(1); v <= 9; v++ {
		hint := uint(1 << (v - 1))
		// cellPeers will contain a list of positions for the given value 'v'
		// and positions visible to it in a container ONLY when that container
		// contains only those two hints
		cellPeers := make(map[int][]int)
		for r := 0; r < 9; r++ {
			for c := 0; c < 9; c++ {
				pos := r*9 + c
				if b.solved[pos] != 0 {
					continue
				}
				if b.blits[pos]&hint == 0 {
					continue
				}

				var links []int
				getSingleLink := func(target int, source int) error {
					if target == source {
						return nil
					}
					if b.solved[target] != 0 {
						return nil
					}
					if b.blits[target]&hint != hint {
						return nil
					}

					links = append(links, target)
					return nil
				}

				allLinks := make(map[int]interface{})

				// row
				if err = b.operateOnRow(pos, getSingleLink); err != nil {
					return err
				}
				if len(links) == 1 {
					for _, item := range links {
						allLinks[item] = struct{}{}
					}
				}

				// column
				links = make([]int, 0)
				if err = b.operateOnColumn(pos, getSingleLink); err != nil {
					return err
				}
				if len(links) == 1 {
					for _, item := range links {
						allLinks[item] = struct{}{}
					}
				}

				// box
				links = make([]int, 0)
				if err = b.operateOnBox(pos, getSingleLink); err != nil {
					return err
				}
				if len(links) == 1 {
					for _, item := range links {
						allLinks[item] = struct{}{}
					}
				} else if len(links) > 1 {
					// delete links if there's more than 1 in a box
					for _, item := range links {
						if _, ok := allLinks[item]; ok {
							delete(allLinks, item)
						}
					}
				}

				if len(allLinks) != 0 {
					links = make([]int, 0)
					for k := range allLinks {
						links = append(links, k)
					}
					sort.Ints(links)
					cellPeers[pos] = links
				}
			}
		}

		// we need to consider only contiguous chains
		// it's possible to have two distinct chains with the same hint
		for len(cellPeers) != 0 {
			posColor := make(map[int]int)
			i := 0
			for k, v := range cellPeers {
				color, ok := posColor[k]
				if !ok {
					if i != 0 {
						continue
					}
					color = 0
					posColor[k] = color
				}
				i++

				flippedColor := 1 - color

				for _, peer := range v {
					peerColor, peerOK := posColor[peer]
					if peerOK {
						if peerColor != flippedColor {
							// contradiction
							continue valueLoop
						}
					} else {
						posColor[peer] = flippedColor
					}
				}
			}

			var color0 []int
			var color1 []int
			for k, color := range posColor {
				delete(cellPeers, k)
				if color == 0 {
					color0 = append(color0, k)
				} else {
					color1 = append(color1, k)
				}
			}

			if len(color0) != len(color1) {
				continue
			}

			for _, pos0 := range color0 {
				for _, pos1 := range color1 {
					vis0 := b.getVisibleCellsWithHint(pos0, hint)
					vis1 := b.getVisibleCellsWithHint(pos1, hint)

					both := intersect(vis0, vis1)
					both = subtract(both, color0)
					both = subtract(both, color1)

					if len(both) > 0 {
						for _, elem := range both {
							var logEntry *updateLog
							if logEntry, err = b.updateCandidates(elem, ^hint); err != nil {
								return err
							}

							if logEntry != nil {
								var args []interface{}
								for k := range posColor {
									args = append(args, k)
								}
								args = append(args, pos0)
								args = append(args, pos1)
								args = append(args, hint)
								b.AddLog(technique, logEntry, "chain="+strings.Repeat("%v ", len(posColor))+" color1=%v color2=%v hint=%v", args...)
							}
						}
						return nil
					}
				}
			}
		}
	}
	return nil
}
