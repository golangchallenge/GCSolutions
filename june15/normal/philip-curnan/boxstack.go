package main

import (
	"fmt"
	"os"
	"sync"
)


// boxStack contains a stack that holds box id numbers, and a mutex to control access.
type boxStack struct {
	lock  sync.Mutex
	stack []uint32
	// quick read count that must not be depended on when allocating boxes
	count int
}

// pop removes a boxid from the stack, returns the boxid and a flag that pop is valid
func (bs *boxStack) pop() (boxid uint32, notEmpty bool) {
	defer func() {
		if x := recover(); x != nil {
			fmt.Printf("panic in boxStack.pop: error: %v\n", x)
			os.Exit(1)
		}
	}()
	boxid = 0
	bs.lock.Lock()
	// make sure the stack is not empty
	notEmpty = len(bs.stack) > 0
	if notEmpty {
		boxid, bs.stack = bs.stack[len(bs.stack)-1], bs.stack[:len(bs.stack)-1]
		bs.count = len(bs.stack)
	}
	bs.lock.Unlock()
	return boxid, notEmpty
}

// push adds a boxid to the stack
func (bs *boxStack) push(x uint32) {
	defer func() {
		if x := recover(); x != nil {
			fmt.Printf("panic in boxStack.push: error: %v\n", x)
			os.Exit(1)
		}
	}()
	bs.lock.Lock()
	// increase the capacity of the stack if it is out of space
	if len(bs.stack) == cap(bs.stack) {
		newbuff := make([]uint32, 0, 2 * cap(bs.stack))
		newbuff = append(newbuff, bs.stack...)
		bs.stack = newbuff
	}
	bs.stack = append(bs.stack, x)
	bs.count = len(bs.stack)
	bs.lock.Unlock()
	return
}
