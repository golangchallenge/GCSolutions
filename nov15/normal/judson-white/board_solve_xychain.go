package main

import (
	"strings"
)

func (b *board) SolveXYChain() error {
	// http://www.sudokuwiki.org/XY_Chains
	// bi-value cells linked together by one value (and visible to each other)
	// terminate when more than one away and cell shares a value with the other end
	// cells visible by both ends of the chain can have their shared value removed.
	for i := 0; i < 81; i++ {
		blit := b.blits[i]
		if GetNumberOfSetBits(blit) != 2 {
			continue
		}

		bits := GetBitList(blit)
		for _, bit := range bits {
			updated, err := b.xyChainTestPosition(i, bit)
			if err != nil {
				return err
			}
			if updated {
				// let simpler techniques take over
				return nil
			}
		}
	}
	return nil
}

func (b *board) xyChainTestPosition(i int, excludeBit uint) (bool, error) {
	const technique = "XY-CHAIN"

	hint := excludeBit
	lists := b.xyChainFollow([]int{i}, excludeBit, hint, 1)
	for _, list := range lists {
		startPos := list[0]
		endPos := list[len(list)-1]
		visible1 := b.getVisibleCells(startPos)
		visible2 := b.getVisibleCells(endPos)
		targets := intersect(visible1, visible2)
		if len(targets) == 0 {
			continue
		}

		updated := false
	targetLoop:
		for _, target := range targets {
			// items in the chain aren't candidates (but why not? shouldn't the logic hold? TODO)
			for _, chainItem := range list {
				if target == chainItem {
					continue targetLoop
				}
			}

			targetBlit := b.blits[target]
			if targetBlit&hint == hint {
				logEntry, err := b.updateCandidates(target, ^hint)
				if err != nil {
					return false, err
				}

				if logEntry != nil {
					var args []interface{}
					for _, chainItem := range list {
						args = append(args, chainItem)
					}
					args = append(args, hint)
					b.AddLog(technique, logEntry, strings.Repeat("%v ", len(list))+"hint %v", args...)
				}

				updated = b.changed
			}
		}

		if updated {
			// let simpler techniques take over
			return true, nil
		}
	}
	return false, nil
}

func (b *board) xyChainFollow(chain []int, excludeBit uint, firstBitInChain uint, depth int) [][]int {
	var lists [][]int

	curPos := chain[len(chain)-1]
	curBlit := b.blits[curPos]

	visible := b.getVisibleCells(curPos)

	var filtered []int
loopVisible:
	for _, item := range visible {
		// avoid cycles
		for _, prevItem := range chain {
			if prevItem == item {
				continue loopVisible
			}
		}
		// ensure cell has 2 hints and is linked to the previous cell
		itemBlit := b.blits[item]
		if GetNumberOfSetBits(itemBlit) != 2 || GetNumberOfSetBits(curBlit&itemBlit) == 0 {
			continue
		}
		if curBlit&itemBlit == excludeBit {
			continue
		}
		filtered = append(filtered, item)
	}

	if len(filtered) == 0 {
		return lists
	}

	for _, item := range filtered {
		itemBlit := b.blits[item]

		var newChain []int
		newChain = append(newChain, chain...)
		newChain = append(newChain, item)

		nextExcludeBit := curBlit & itemBlit &^ excludeBit

		if len(chain) > 1 {
			if itemBlit&^nextExcludeBit&firstBitInChain == firstBitInChain {
				// TODO: should we keep going? there may be longer chains
				lists = append(lists, newChain)
				continue
			}
		}

		newLists := b.xyChainFollow(newChain, nextExcludeBit, firstBitInChain, depth+1)
		lists = append(lists, newLists...)
	}

	return lists
}
