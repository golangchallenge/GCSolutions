package main

import "testing"

func TestUtils(t *testing.T) {
	testBox := box{x: 1, y: 2, w: 1, l: 2, id: 2}
	outBox := sideWays(testBox)
	if outBox.l > outBox.w {
		t.Error("Expected outbox's length to be shorter than or equal to width. Didn't get it")
	}
	outBox = upRight(outBox)
	if outBox.w > outBox.l {
		t.Error("Upright boxes's length has to exceed the width. Condition failed.")
	}
}
