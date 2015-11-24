package main

import (
	"strconv"
)

// HasSingleBit returns true if the specified uint has only 1 bit set.
func HasSingleBit(val uint) bool {
	if val == 0 {
		return false
	}
	return val&(val-1) == 0
}

// GetNumberOfSetBits returns the number of set bits in the specified uint.
func GetNumberOfSetBits(val uint) uint {
	var count uint
	for count = 0; val != 0; count++ {
		val &= val - 1
	}
	return count
}

// GetSingleBitValue returns the position (starting at 1) of a single set bit.
// Note GetSingleBitValue does not test only a single bit is set; it's assumed
// the caller has already validated the input with HasSingleBit or is using
// the results of GetBitList.
//
// In terms of Sudoku this is useful for getting the last remaining hint of
// a cell where all hints are encoded in a single uint.
func GetSingleBitValue(val uint) uint {
	var m uint
	for m = 1; m <= 9; m++ {
		if val == 1 {
			break
		}
		val >>= 1
	}
	return m
}

// GetBitsString returns a comma separated list of set bit positions (starting at 1).
// In terms of Sudoku this is useful for displaying a list of remaining hints.
func GetBitsString(val uint) string {
	var msg string
	for m := 1; m <= 9; m++ {
		if val&0x01 == 1 {
			if msg != "" {
				msg += ","
			}
			msg += strconv.Itoa(m)
		}
		val >>= 1
	}
	return msg
}

// GetBitList takes a uint and expands the set bits into a slice
// of uint's which each entry having a single bit set.
// In terms of Sudoku this is useful for expanding a list of hints
// into their single bit representations.
func GetBitList(val uint) []uint {
	var list []uint
	for m := uint(1); m <= 9; m++ {
		if val&0x01 == 1 {
			list = append(list, 1<<(m-1))
		}
		val >>= 1
	}
	return list
}
