package main

import (
	"reflect"

	"testing"
)

func assertEqual(t *testing.T, expected, got interface{}) {
	if expected != got {
		t.Fatalf("expected %v; got %v", expected, got)
	}
}

func assertNotEqual(t *testing.T, expected, got interface{}) {
	if expected == got {
		t.Fatalf("expected not equal to %v; got %v", expected, got)
	}
}

func assertTrue(t *testing.T, b bool) {
	if !b {
		t.Fatal("expected true")
	}
}

func assertNil(t *testing.T, i interface{}) {
	v := reflect.ValueOf(i)

	if v.IsValid() && !v.IsNil() {
		t.Fatalf("expected nil, got %v", i)
	}
}

func assertNotNil(t *testing.T, i interface{}) {
	if i == interface{}(nil) {
		t.Fatalf("expected non-nil")
	}
}

func TestSplitNode(t *testing.T) {
	p := newBinaryTreePacker()
	p.subdivide(&p.root, 2, 2)
	assertEqual(t, uint8(4), p.root.down.w)
	assertEqual(t, uint8(2), p.root.down.h)
	assertEqual(t, uint8(2), p.root.right.w)
	assertEqual(t, uint8(2), p.root.right.h)

	p = newBinaryTreePacker()
	p.subdivide(&p.root, 1, 3)
	assertEqual(t, uint8(4), p.root.down.w)
	assertEqual(t, uint8(1), p.root.down.h)
	assertEqual(t, uint8(3), p.root.right.w)
	assertEqual(t, uint8(3), p.root.right.h)

	n, ok := p.search(&p.root, 2, 2)
	assertTrue(t, ok)
	assertNotEqual(t, &p.root, n)
}

func TestFindNode(t *testing.T) {
	p := newBinaryTreePacker()
	n, ok := p.search(&p.root, 2, 2)
	p.subdivide(n, 2, 2)
	assertTrue(t, ok)
	assertTrue(t, n.used)
	assertTrue(t, p.root.used)

	n, ok = p.search(&p.root, 2, 2)
	assertNotEqual(t, &p.root, n)
	assertEqual(t, uint8(2), n.x)
	assertEqual(t, uint8(0), n.y)
	assertEqual(t, uint8(2), n.w)
	assertEqual(t, uint8(2), n.h)

	n, ok = p.search(&p.root, 2, 2)
	assertTrue(t, ok)
	n, ok = p.search(&p.root, 2, 2)
	assertTrue(t, ok)
	// This box can't fit
	n, ok = p.search(&p.root, 1, 3)
	assertTrue(t, !ok)
}

func TestPack(t *testing.T) {
	p := newBinaryTreePacker()
	rest, pallet := p.pack([]box{
		{x: 0, y: 0, w: 1, l: 3},
		{x: 0, y: 0, w: 1, l: 3},
	})
	assertEqual(t, 0, len(rest))
	assertEqual(t, 2, len(pallet.boxes))

	p = newBinaryTreePacker()
	rest, pallet = p.pack([]box{
		{x: 0, y: 0, w: 1, l: 3},
		{x: 0, y: 0, w: 3, l: 3},
		{x: 0, y: 0, w: 1, l: 1},
		{x: 0, y: 0, w: 2, l: 1},
	})
	assertEqual(t, 0, len(rest))
	assertEqual(t, 4, len(pallet.boxes))
	assertNil(t, pallet.IsValid())
}
