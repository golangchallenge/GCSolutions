package main

import (
	"fmt"
	"testing"
)

// NOTE: 'expected' values assume palletWidth and palletLength = 4

func TestCheckCellValid(t *testing.T) {

	// valid cell
	actualErr := checkCellValid(0, 0)
	expectedErr := error(nil)
	if actualErr != expectedErr {
		t.Errorf("checkCellValid(0,0) returned err:%v; expected:%v", actualErr, expectedErr)
	}

	// valid cell
	actualErr = checkCellValid(palletLength-1, palletWidth-1)
	expectedErr = error(nil)
	if actualErr != expectedErr {
		t.Errorf("checkCellValid(palletLength-1, palletWidth-1) returned err:%v; expected:%v", actualErr, expectedErr)
	}

	actualErr = checkCellValid(0, palletWidth)
	expectedErr = errBadRow(palletWidth)
	if actualErr != expectedErr {
		t.Errorf("checkCellValid(0, palletWidth) returned err:%v; expected:%v", actualErr, expectedErr)
	}

	actualErr = checkCellValid(palletLength, 0)
	expectedErr = errBadCol(palletLength)
	if actualErr != expectedErr {
		t.Errorf("checkCellValid(palletLength, 0) returned err:%v; expected:%v", actualErr, expectedErr)
	}
}

func TestPackPlan(t *testing.T) {
	plan := new(packPlan)
	var actualErr, expectedErr error
	actualPlan := plan.String()
	expectedPlan := `
____
____
____
____
`
	if actualPlan != expectedPlan {
		t.Errorf("empty packPlan returns &v expected %v", actualPlan, expectedPlan)
	}

	// paint first box
	plan.markBox(0, 0, 3)
	actualPlan = plan.String()
	expectedPlan = `
#___
#___
#___
____
`
	if actualPlan != expectedPlan {
		t.Errorf("packPlan.markBox(0, 0, 3) results in %v expected %v", actualPlan, expectedPlan)
	}

	// cell not empty
	actualErr = plan.markBox(0, 2, 2)
	expectedErr = errCellNotEmpty
	if actualErr != expectedErr {
		t.Errorf("packPlan.markBox(0, 2, 2) returned err:%v; expected:%v", actualErr, expectedErr)
	}

	// check if plan changed
	actualPlan = plan.String()
	if actualPlan != expectedPlan {
		t.Errorf("packPlan.markBox(0, 2, 2) results in %v expected %v", actualPlan, expectedPlan)
	}

	// bad column value
	actualErr = plan.markBox(palletLength, 0, 4)
	expectedErr = errBadCol(palletLength)
	if actualErr != expectedErr {
		t.Errorf("packPlan.markBox(palletLength, 0, 4) returned err:%v; expected:%v", actualErr, expectedErr)
	}

	// check if plan changed
	actualPlan = plan.String()
	if actualPlan != expectedPlan {
		t.Errorf("packPlan.markBox(palletLength, 0, 4) results in %v expected %v", actualPlan, expectedPlan)
	}

	// bad row value
	actualErr = plan.markBox(2, palletWidth, 4)
	expectedErr = errBadRow(palletWidth)
	if actualErr != expectedErr {
		t.Errorf("packPlan.markBox(2, palletWidth, 4) returned err:%v; expected:%v", actualErr, expectedErr)
	}

	// check if plan changed
	actualPlan = plan.String()
	if actualPlan != expectedPlan {
		t.Errorf("packPlan.markBox(2, palletWidth, 4) results in %v expected %v", actualPlan, expectedPlan)
	}

	// paint second box
	actualErr = plan.markBox(2, 0, 4)
	expectedErr = error(nil)
	if actualErr != expectedErr {
		t.Errorf("packPlan.markBox(2, 0, 4) returned err:%v; expected:%v", actualErr, expectedErr)
	}

	// paint third box
	actualErr = plan.markBox(3, 0, 2)
	expectedErr = error(nil)
	if actualErr != expectedErr {
		t.Errorf("packPlan.markBox(2, 0, 4) returned err:%v; expected:%v", actualErr, expectedErr)
	}

	actualPlan = plan.String()
	expectedPlan = `
#_##
#_##
#_#_
__#_
`
	if actualPlan != expectedPlan {
		t.Errorf("packPlan.markBox() after adding two boxes results in %v expected %v", actualPlan, expectedPlan)
	}

	// cell is not empty
	l, w := plan.getAvailSpace(openCell{0, 0})
	expectedL, expectedW := uint8(0), uint8(0)
	if l != expectedL || w != expectedW {
		t.Errorf("packPlan.getAvailSpace(openCell{0, 0}) returned l:%v, w:%v; expected l:%v, w:%v", l, w, expectedL, expectedW)
	}

	// cell is empty
	l, w = plan.getAvailSpace(openCell{0, 3})
	expectedL, expectedW = 2, 1
	if l != expectedL || w != expectedW {
		t.Errorf("packPlan.getAvailSpace(openCell{0, 3}) returned l:%v, w:%v; expected l:%v, w:%v", l, w, expectedL, expectedW)
	}

	// cell is invalid
	l, w = plan.getAvailSpace(openCell{palletLength, palletWidth})
	expectedL, expectedW = 0, 0
	if l != expectedL || w != expectedW {
		t.Errorf("packPlan.getAvailSpace(openCell{palletLength, palletWidth}) returned l:%v, w:%v; expected l:%v, w:%v", l, w, expectedL, expectedW)
	}

	plan.clear()
	actualPlan = plan.String()
	expectedPlan = `
____
____
____
____
`
	if actualPlan != expectedPlan {
		t.Errorf("packPlan.clear() results in %v expected %v", actualPlan, expectedPlan)
	}

}

func TestOpenCellStack(t *testing.T) {
	var cell openCell
	var stack openCellStack
	var actualErr, expectedErr error

	// new stack
	expectedLength := 0
	if len(stack) != expectedLength {
		t.Errorf("new openCellStack length is %d; expected %d", len(stack), expectedLength)
	}

	// push valid cell
	stack.push(openCell{1, 3})
	expectedLength = 1
	if len(stack) != expectedLength {
		t.Errorf("openCellStack.push(openCell{1, 3}) results in stack length of %d; expected %d", len(stack), expectedLength)
	}

	// push invalid cell
	stack.push(openCell{palletLength, palletWidth})
	expectedLength = 1
	if len(stack) != expectedLength {
		t.Errorf("openCellStack.push(openCell{palletLength, palletWidth}) results in stack length of %d; expected %d", len(stack), expectedLength)
	}

	// valid peek
	cell, actualErr = stack.peek()
	expectedErr = nil
	expectedLength = 1
	expectedCol := 1
	expectedRow := 3
	if actualErr != expectedErr {
		t.Errorf("openCellStack.peek() returned error:%v; expected:%v", actualErr, expectedErr)
	}
	if len(stack) != expectedLength {
		t.Errorf("openCellStack.peek() results in stack length of %d; expected %d", len(stack), expectedLength)
	}
	if int(cell.col) != expectedCol || int(cell.row) != expectedRow {
		t.Errorf("openCellStack.peek() returned a cell with location (%d,%d); expected (%d,%d)", cell.col, cell.row, expectedCol, expectedRow)
	}

	// valid pop
	cell, actualErr = stack.pop()
	expectedErr = nil
	expectedLength = 0
	expectedCol = 1
	expectedRow = 3
	if actualErr != expectedErr {
		t.Errorf("openCellStack.pop() returned error:%v; expected:%v", actualErr, expectedErr)
	}
	if len(stack) != expectedLength {
		t.Errorf("openCellStack.pop() results in stack length of %d; expected %d", len(stack), expectedLength)
	}
	if int(cell.col) != expectedCol || int(cell.row) != expectedRow {
		t.Errorf("openCellStack.pop() returned a cell with location (%d,%d); expected location (%d,%d)", cell.col, cell.row, expectedCol, expectedRow)
	}

	// valid push
	stack.push(openCell{0, 0})
	expectedLength = 1
	if len(stack) != expectedLength {
		t.Errorf("openCellStack.push(openCell{0, 0}) results in stack length of %d; expected %d", len(stack), expectedLength)
	}

	// clear stack
	stack.clear()
	expectedLength = 0
	if len(stack) != expectedLength {
		t.Errorf("openCellStack.clear() results in stack length of %d; expected %d", len(stack), expectedLength)
	}

	// peek while stack is empty
	cell, actualErr = stack.peek()
	expectedErr = errStackEmpty
	expectedLength = 0
	expectedCol = palletLength
	expectedRow = palletWidth
	if actualErr != expectedErr {
		t.Errorf("openCellStack.peek() returned err:%v; expected:%v", actualErr, expectedErr)
	}
	if len(stack) != expectedLength {
		t.Errorf("openCellStack.peek() results in stack length of %d; expected %d", len(stack), expectedLength)
	}
	if int(cell.col) != expectedCol || int(cell.row) != expectedRow {
		t.Errorf("openCellStack.peek() returned a cell with location (%d,%d); expected (%d,%d)", cell.col, cell.row, expectedCol, expectedRow)
	}

	// pop while stack is empty
	cell, actualErr = stack.pop()
	expectedErr = errStackEmpty
	expectedLength = 0
	expectedCol = palletLength
	expectedRow = palletWidth
	if actualErr != expectedErr {
		t.Errorf("openCellStack.pop() returned err:%v; expected:%v", actualErr, expectedErr)
	}
	if len(stack) != expectedLength {
		t.Errorf("openCellStack.pop() results in stack length of %d; expected %d", len(stack), expectedLength)
	}
	if int(cell.col) != expectedCol || int(cell.row) != expectedRow {
		t.Errorf("openCellStack.pop() returned a cell with location (%d,%d); expected location (%d,%d)", cell.col, cell.row, expectedCol, expectedRow)
	}
}

func TestCargoQueue(t *testing.T) {
	actualCandidate := 0
	expectedCandidate := 0
	ok := false
	q := make(cargoQueue, 0)
	expectedLength := 0
	expectedString := ""
	actualString := ""

	if len(q) != expectedLength {
		t.Error("new cargoQueue has length %v; expected %v", len(q), expectedLength)
	}

	q.addCargo(cargo{item: box{id: 1, x: 0, y: 0, l: 2, w: 1}, area: 2})
	expectedString = "[<ID: 1, 2 x 1, 2 area>]"
	actualString = fmt.Sprintf("%v", q)
	if actualString != expectedString {
		t.Error("cargoQueue.addCargo() produces queue %v, expected %v", actualString, expectedString)
	}

	q.addCargo(cargo{item: box{id: 2, x: 0, y: 0, l: 3, w: 3}, area: 9})
	expectedString = "[<ID: 1, 2 x 1, 2 area> <ID: 2, 3 x 3, 9 area>]"
	actualString = fmt.Sprintf("%v", q)
	if actualString != expectedString {
		t.Error("cargoQueue.addCargo() produces queue %v, expected %v", actualString, expectedString)
	}

	q.addCargo(cargo{item: box{id: 3, x: 0, y: 0, l: 2, w: 2}, area: 4})
	expectedString = "[<ID: 1, 2 x 1, 2 area> <ID: 2, 3 x 3, 9 area> <ID: 3, 2 x 2, 4 area>]"
	actualString = fmt.Sprintf("%v", q)
	if actualString != expectedString {
		t.Error("cargoQueue.addCargo() produces queue %v, expected %v", actualString, expectedString)
	}

	q.sort()
	expectedString = "[<ID: 2, 3 x 3, 9 area> <ID: 3, 2 x 2, 4 area> <ID: 1, 2 x 1, 2 area>]"
	actualString = fmt.Sprintf("%v", q)
	if actualString != expectedString {
		t.Error("cargoQueue.sort() produces queue %v, expected %v", actualString, expectedString)
	}

	expectedLength = 3
	if len(q) != expectedLength {
		t.Errorf("cargoQueue has length %v; expected %v", len(q), expectedLength)
	}

	actualCandidate, ok = q.findNextCandidate(4, 0)
	expectedCandidate = 1
	if !ok {
		t.Error("cargoQueue.findNextCandidate(4,0) should have been successful.")
	}
	if actualCandidate != expectedCandidate {
		t.Error("cargoQueue.findNextCandidate(4,0) returned index %v; expected %v", actualCandidate, expectedCandidate)
	}

	q.removeCargo(actualCandidate)
	actualString = fmt.Sprintf("%v", q)
	expectedString = "[<ID: 2, 3 x 3, 9 area> <ID: 1, 2 x 1, 2 area>]"
	if actualString != expectedString {
		t.Error("cargoQueue.removeCargo() produces queue %v, expected %v", actualString, expectedString)
	}

	q.removeCargo(0)
	actualString = fmt.Sprintf("%v", q)
	expectedString = "[<ID: 1, 2 x 1, 2 area>]"
	if actualString != expectedString {
		t.Error("cargoQueue.removeCargo() produces queue %v, expected %v", actualString, expectedString)
	}

	expectedLength = 1
	if len(q) != expectedLength {
		t.Errorf("cargoQueue has length %v; expected %v", len(q), expectedLength)
	}
}

func TestTurnVertical(t *testing.T) {
	bx := box{x: 0, y: 0, w: 4, l: 2, id: 99}
	testBox := bx

	expectedW, expectedL := bx.w, bx.l
	if testBox.turnVertical(); testBox.w != expectedW || testBox.l != expectedL {
		t.Errorf("turnVertical(%v) resulted in w:%v x l:%v, wanted w:%v x l:%v", bx, testBox.w, testBox.l, expectedW, expectedL)
	}

	bx = box{x: 0, y: 0, w: 2, l: 4, id: 99}
	testBox = bx

	expectedW, expectedL = bx.l, bx.w
	if testBox.turnVertical(); testBox.w != expectedW || testBox.l != expectedL {
		t.Errorf("turnVertical(%v) resulted in w:%v x l:%v, wanted w:%v x l:%v", bx, testBox.w, testBox.l, expectedW, expectedL)
	}

	bx = box{x: 0, y: 0, w: 3, l: 3, id: 99}
	testBox = bx

	expectedW, expectedL = bx.w, bx.l
	if testBox.turnVertical(); testBox.w != expectedW || testBox.l != expectedL {
		t.Errorf("turnVertical(%v) resulted in w:%v x l:%v, wanted w:%v x l:%v", bx, testBox.w, testBox.l, expectedW, expectedL)
	}
}
