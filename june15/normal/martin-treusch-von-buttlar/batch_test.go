package main

import (
	"io"
	"os"
	"testing"
)

func loadTrucksFromFixture(fname string, maxTrucks int) (*Batch, error) {
	f, err := os.Open("testdata/" + fname)
	if err != nil {
		return nil, err
	}

	tr := newTruckReader(f)
	if err != nil {
		return nil, err
	}

	b := NewBatch()
	i := 0
	for {
		truck, err := tr.Next()
		if err != nil {
			if err != io.EOF {
				return nil, err
			}
			break
		}
		b.UnpackTruck(truck)
		i++
		if maxTrucks > 0 && i == maxTrucks {
			break
		}
	}

	return b, nil
}

func TestRepackEmptyBatch(t *testing.T) {
	b := NewBatch()
	b.pallets = make([]pallet, b.openPallets)
	packedBoxes := b.RepackPallets()
	if packedBoxes != 0 {
		t.Errorf("Expected 0 packed boxes, got %d", packedBoxes)
	}
}

func TestSendEmptyBatch(t *testing.T) {
	b := NewBatch()
	b.pallets = make([]pallet, b.openPallets)
	out := make(chan *truck, 1)
	sentTrucks := b.SendTrucks(out)
	if sentTrucks != 0 {
		t.Errorf("Expected 0 sent trucks, got %d", sentTrucks)
	}
}

func TestFFDPacker(t *testing.T) {
	testPacker(t, repackFFD)
}

func TestMFFDPacker(t *testing.T) {
	testPacker(t, repackMFFD)
}

func testPacker(t *testing.T, packer func(*Batch) int) {
	pallets := make([]pallet, 20)
	boxes := make(map[box]bool)
	for i := 0; i < len(pallets); i++ {
		pallets[i] = genpal()
		for _, box := range pallets[i].boxes {
			boxes[box.canon()] = true
		}
	}

	b := NewBatch()
	b.repacker = packer
	b.UnpackTruck(&truck{id: 1, pallets: pallets})
	b.pallets = make([]pallet, b.openPallets)
	packedBoxes := b.RepackPallets()
	if packedBoxes != len(boxes) {
		t.Errorf("expected %d packed boxes, got %d", len(boxes), packedBoxes)
	}
	for _, p := range b.pallets {
		for _, box := range p.boxes {
			if _, e := boxes[box.canon()]; e {
				delete(boxes, box.canon())
			} else {
				t.Errorf("new box: %s", box)
			}
		}
	}
	if len(boxes) > 0 {
		for box := range boxes {
			t.Errorf("missing box: %s", box)
		}
	}
}

func TestUnpackTruck(t *testing.T) {
	fixtures := []struct {
		fname           string
		packer          func(*Batch) int
		trucks2load     int
		loadedPallets   int
		boxesPacked     int
		fullPallets     int
		halfFullPallets int
	}{
		{"100trucks.txt", repackFFD, 2, 19, 36, 10, 1},
		{"100trucks.txt", repackMFFD, 2, 19, 36, 9, 2},
		{"100trucks.txt", repackFFD, -1, 944, 1815, 618, 2},
		{"100trucks.txt", repackMFFD, -1, 944, 1815, 618, 2},
		//{"300trucks.txt", repackFFD, 200, 1907, 3684, 1258, 1},
	}

	for _, f := range fixtures {
		b, err := loadTrucksFromFixture(f.fname, f.trucks2load)
		if err != nil {
			t.Error(err)
			continue
		}
		b.repacker = f.packer

		if b.openPallets != f.loadedPallets {
			t.Errorf("expected %d pallets, got %d.", f.loadedPallets, b.openPallets)
		}

		b.pallets = make([]pallet, b.openPallets)
		boxesPacked := b.RepackPallets()
		if boxesPacked != f.boxesPacked {
			t.Errorf("expected %d packed boxes, got %d.", f.boxesPacked, boxesPacked)
		}
		fullPallets := 0
		halfFullPallets := 0
		for _, p := range b.pallets {
			if p.Items() == 0 {
				continue
			}
			if p.IsFull() {
				fullPallets++
			} else {
				halfFullPallets++
			}
		}
		if fullPallets != f.fullPallets {
			t.Errorf("expected %d full boxes, got %d.", f.fullPallets, fullPallets)
		}
		if halfFullPallets != f.halfFullPallets {
			t.Errorf("expected %d half filled boxes, got %d.", f.halfFullPallets, halfFullPallets)
		}
	}
}

func BenchmarkRepackFFD10(b *testing.B) {
	benchmarkRepack("500trucks.txt", b, 10, repackFFD)
}

func BenchmarkRepackMFFD10(b *testing.B) {
	benchmarkRepack("500trucks.txt", b, 10, repackMFFD)
}

func BenchmarkRepackFFD250(b *testing.B) {
	benchmarkRepack("500trucks.txt", b, 250, repackFFD)
}

func BenchmarkRepackMFFD250(b *testing.B) {
	benchmarkRepack("500trucks.txt", b, 250, repackMFFD)
}

func benchmarkRepack(fname string, b *testing.B, maxCount int, r func(*Batch) int) {
	batch, err := loadTrucksFromFixture(fname, maxCount)
	batch.repacker = r
	if err != nil {
		panic(err)
	}

	for n := 0; n < b.N; n++ {
		batch.pallets = make([]pallet, batch.openPallets)
		batch.RepackPallets()
	}
}
