package main

import (
	"testing"
	"time"
)

// TestEmptyPushPop tests that boxStack knows when its empty and can push and pop
func TestEmptyPushPop(t *testing.T) {
	rp := &repacker{}
	// new stack should be empty
	_, notEmpty := rp.boxes[0][0].pop()
	if notEmpty {
		t.Fatalf("boxStack is empty but says it's not empty")
	}
	// empty stack should have count 0
	if rp.boxes[0][0].count != 0 {
		t.Fatalf("boxStack count is %d when it should be 0", rp.boxes[0][0].count)
	}
	// push a box onto stack
	pushid := uint32(100)
	rp.boxes[0][0].push(pushid)
	if rp.boxes[0][0].count != 1 {
		t.Fatalf("boxStack count is %d when it should be 1", rp.boxes[0][0].count)
	}
	// pop box off of stack
	popid, notEmpty := rp.boxes[0][0].pop()
	if !notEmpty {
		t.Fatalf("boxStack is empty again but says its not empty")
	}
	// pushed and popped id must be the same
	if popid != pushid {
		t.Fatalf("popped box id != pushed box id ")
	}
	if rp.boxes[0][0].count != 0 {
		t.Fatalf("boxStack count is %d when it should be 0", rp.boxes[0][0].count)
	}
}

// TestLocking tests boxStack boxes is protected by mutex
func TestLocking(t *testing.T) {
	rp := &repacker{}

	// lock and hold in a go routine, test no one snuck in during wait
	go func() {
		rp.boxes[0][0].lock.Lock()
		time.Sleep(100 * time.Millisecond)
		if len(rp.boxes[0][0].stack) != 0 {
			t.Fatalf("boxStack length is %d when it should be 0", len(rp.boxes[0][0].stack))
		}
		rp.boxes[0][0].lock.Unlock()
	}()
	// try to get in before lock is released
	rp.boxes[0][0].push(100)
}
