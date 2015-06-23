package main

import (
	"io"
	"os"
	"strings"
	"testing"
)

const testTruck = `truck 1
0 0 1 1 101,1 1 1 1 102,2 2 1 1 103,3 0 4 1 104
0 0 1 1 101,0 0 1 1 102
0 0 5 5 101
endtruck
`

func TestTruckReader(t *testing.T) {
	r := newTruckReader(strings.NewReader(testTruck))

	truck, err := r.Next()
	if err != nil {
		t.Fatal("truck read:", err)
	}

	if truck.id != 1 {
		t.Fatalf("truck id %v, expected 1", truck.id)
	}
	expPallets := 3
	if len(truck.pallets) != expPallets {
		t.Fatalf("truck has %v pallets, expected %v", len(truck.pallets), expPallets)
	}

	// Test String() formatting.
	expected := `
| !       |
|   @     |
|     #   |
| $ $ $ $ |
`
	s := truck.pallets[0].String()
	if s != expected {
		t.Error("pallet 0 format is wrong:", s)
	}
	t.Log(s)
	s = truck.pallets[0].boxes[0].String()
	expected = "0 0 1 1 101"
	if s != expected {
		t.Error("pallet 0 box 0 format is wrong:", s)
	}

	_, err = truck.pallets[1].paint()
	if _, ok := err.(errOverlap); !ok {
		t.Error("pallet 1 error is wrong:", err)
	}

	_, err = truck.pallets[2].paint()
	if _, ok := err.(errEdge); !ok {
		t.Error("pallet 2 error is wrong:", err)
	}
}

func TestNoInputTruckReader(t *testing.T) {
	r := newTruckReader(strings.NewReader(""))
	_, err := r.Next()
	if err != io.EOF {
		t.Error("expected eof, got:", err)
	}
}

type errReader struct{}

func (er errReader) Read(buf []byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}

func TestErrTruckReader(t *testing.T) {
	r := newTruckReader(errReader{})
	_, err := r.Next()
	if err == nil {
		t.Error("missing error")
	}
	_, err = r.Next()
	if err == nil {
		t.Error("missing 2nd error")
	}
}

func TestBadPallet(t *testing.T) {
	// gridbox missing id
	_, err := palletFromString("1 1 5 5")
	if err != io.EOF {
		t.Error("wrong err:", err)
	}

	// pallet with a comma on the end
	_, err = palletFromString("1 1 5 5 101,")
	if err != errEmpty {
		t.Error("wrong err:", err)
	}

	// zero sized box
	_, err = palletFromString("1 1 0 0 101")
	if err != errZeroBox {
		t.Error("wrong err:", err)
	}
}

func BenchmarkRead(b *testing.B) {
	f, err := os.Open("testdata/100trucks.txt")
	if err != nil {
		b.Error(err)
	}
	defer f.Close()

	r := newTruckReader(f)
	for {
		_, err := r.Next()
		if err == io.EOF {
			return
		}
		if err != nil {
			b.Error(err)
		}
	}
}
