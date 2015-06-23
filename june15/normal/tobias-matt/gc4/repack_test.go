package main

import "testing"

// func prepareBox(a area, b box) box
func TestPrepareBox(t *testing.T) {
	a := area{x: 1, y: 1, w: 3, l: 1}
	b := box{id: 0, x: 0, y: 0, w: 3, l: 1}
	expected := box{id: 0, x: 1, y: 1, w: 3, l: 1}
	b = prepareBox(a, b)
	if b != expected {
		t.Errorf("box %v, expected %v", b, expected)
	}

	a = area{x: 1, y: 1, w: 1, l: 3}
	b = box{id: 0, x: 0, y: 0, w: 3, l: 1}
	expected = box{id: 0, x: 1, y: 1, w: 1, l: 3}
	b = prepareBox(a, b)
	if b != expected {
		t.Errorf("box %v, expected %v", b, expected)
	}
}

// func (ps *palletSpace) freeAreas() []area
// func (ps *palletSpace) setUsed(b box)
func TestPalletSpace(t *testing.T) {
	ps := palletSpace{}
	b := box{x: 0, y: 0, w: 4, l: 1}
	ps.setUsed(b)
	expected := area{x: 1, y: 0, w: 4, l: 3}
	freeAreas := ps.freeAreas()
	if len(freeAreas) != 1 {
		t.Errorf("freeAreas %v, expected 1", len(freeAreas))
		return
	}
	if freeAreas[0] != expected {
		t.Errorf("area %v, expected %v", freeAreas[0], expected)
	}

	b = box{x: 1, y: 0, w: 1, l: 3}
	ps.setUsed(b)
	expected = area{x: 1, y: 1, w: 3, l: 3}
	freeAreas = ps.freeAreas()
	if len(freeAreas) != 1 {
		t.Errorf("freeAreas %v, expected 1", len(freeAreas))
		return
	}
	if freeAreas[0] != expected {
		t.Errorf("area %v, expected %v", freeAreas[0], expected)
	}
}

// func (r *repacker) Boxes() uint32
// func (r *repacker) PushBox(b box)
// func (r *repacker) PopBox(w, l uint8) (box, error)
func TestBoxes(t *testing.T) {
	r := repacker{}

	b1 := box{id: 1, w: 4, l: 1}
	r.PushBox(b1)

	b2 := box{id: 2, w: 3, l: 3}
	r.PushBox(b2)

	b3 := box{id: 3, w: 1, l: 1}
	r.PushBox(b3)

	boxes := r.Boxes()
	if boxes != 3 {
		t.Errorf("boxes %v, expected 3", boxes)
	}

	b, err := r.PopBox(4, 4)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	if b != b1 {
		t.Errorf("box %v, expected %v", b, b1)
	}
	boxes = r.Boxes()
	if boxes != 2 {
		t.Errorf("boxes %v, expected 2", boxes)
	}

	b, err = r.PopBox(4, 2)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	if b != b3 {
		t.Errorf("box %v, expected %v", b, b3)
	}
	boxes = r.Boxes()
	if boxes != 1 {
		t.Errorf("boxes %v, expected 1", boxes)
	}

	b, err = r.PopBox(3, 2)
	if err == nil {
		t.Errorf("missing error")
	}
	boxes = r.Boxes()
	if boxes != 1 {
		t.Errorf("boxes %v, expected 1", boxes)
	}

	b, err = r.PopBox(4, 4)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	if b != b2 {
		t.Errorf("box %v, expected %v", b, b2)
	}
	boxes = r.Boxes()
	if boxes != 0 {
		t.Errorf("boxes %v, expected 0", boxes)
	}
}

// func (r *repacker) PushTruckID(id int)
// func (r *repacker) PopTruckID(last bool) (int, error)
func TestTruckIDs(t *testing.T) {
	r := repacker{}

	r.PushTruckID(1)
	_, err := r.PopTruckID(false)
	if err == nil {
		t.Errorf("missing error")
	}

	r.PushTruckID(2)
	id, err := r.PopTruckID(false)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	if id != 1 {
		t.Errorf("truck id %v, expected 1", id)
	}

	id, err = r.PopTruckID(true)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	if id != 2 {
		t.Errorf("truck id %v, expected 2", id)
	}
}

// func (r *repacker) discharge(in <-chan *truck)
func TestDischarge(t *testing.T) {
	r := repacker{}
	var testTrucks = []truck{
		{1, []pallet{
			{[]box{{id: 1, w: 1, l: 1}}},
			{[]box{{id: 2, w: 2, l: 2}}},
			{[]box{{id: 3, w: 3, l: 3}}},
			{[]box{{id: 4, w: 4, l: 4}}},
		}},
		{0, []pallet{}},
	}

	in := make(chan *truck)
	go func() {
		for x := 0; x < len(testTrucks); x++ {
			in <- &testTrucks[x]
		}
	}()

	r.discharge(in)

	b, err := r.PopBox(4, 4)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	expected := box{id: 4, w: 4, l: 4}
	if b != expected {
		t.Errorf("box %v, expected %v", b, expected)
	}

	b, err = r.PopBox(4, 4)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	expected = box{id: 3, w: 3, l: 3}
	if b != expected {
		t.Errorf("box %v, expected %v", b, expected)
	}

	b, err = r.PopBox(4, 4)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	expected = box{id: 2, w: 2, l: 2}
	if b != expected {
		t.Errorf("box %v, expected %v", b, expected)
	}

	b, err = r.PopBox(4, 4)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	expected = box{id: 1, w: 1, l: 1}
	if b != expected {
		t.Errorf("box %v, expected %v", b, expected)
	}

	_, err = r.PopTruckID(false)
	if err == nil {
		t.Errorf("missing error")
	}

	id, err = r.PopTruckID(true)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	if id != 1 {
		t.Errorf("truck id %v, epected 1", id)
	}
}

// func (r *repacker) repack(out chan<- *truck)
func TestRepack(t *testing.T) {
	r := repacker{}
	b1 := box{id: 1, w: 4, l: 3}
	r.PushBox(b1)
	b2 := box{id: 2, w: 4, l: 1}
	r.PushBox(b2)
	r.PushTruckID(1)

	out := make(chan *truck)
	go r.repack(out)
	truck := <-out

	if truck.id != 1 {
		t.Errorf("truck id %v, epected 1", truck.id)
	}

	if len(truck.pallets) != 1 {
		t.Errorf("pallets %v, epected 1", len(truck.pallets))
		return
	}

	if len(truck.pallets[0].boxes) != 2 {
		t.Errorf("boxes %v, epected 2", len(truck.pallets))
		return
	}

	if truck.pallets[0].boxes[0] != b1 {
		t.Errorf("box %v, epected %v", truck.pallets[0].boxes[0], b1)
	}
	b2.x = 3
	if truck.pallets[0].boxes[1] != b2 {
		t.Errorf("box %v, epected %v", truck.pallets[0].boxes[1], b2)
	}
}
