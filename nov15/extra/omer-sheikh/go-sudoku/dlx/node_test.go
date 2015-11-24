package dlx

import "testing"

func TestSimpleConnections(t *testing.T) {
	root := NewRoot()
	if root.left != root || root.right != root {
		t.Error("Root connections are invalid")
	}
	header := AddHeader(root)
	if root.left != header || header.right != root {
		t.Error("Header connections are invalid")
	}
	if root.right != header || header.left != root {
		t.Error("Header connections are invalid")
	}
	node1 := AddNode(1, header)
	node2 := AddNode(2, header)
	if header.down != node1 || header.up != node2 || node1.header != header {
		t.Error("Header-Cell connections are invalid")
	}
	if node1.down != node2 || node2.up != node1 {
		t.Error("Cell connections are invalid")
	}
}

func TestBuildRow(t *testing.T) {
	root := NewRoot()
	headers := make([]*Node, 4)
	row := make([]*Node, 4)
	for i := 0; i < 4; i++ {
		headers[i] = AddHeader(root)
		row[i] = AddNode(0, headers[i])
	}
	BuildRow(row)
	for i := range row {
		iLeft := i - 1
		if iLeft < 0 {
			iLeft = len(row) - 1
		}
		iRight := (i + 1) % len(row)
		if row[i].left != row[iLeft] || row[i].right != row[iRight] {
			t.Fail()
		}
	}
}
