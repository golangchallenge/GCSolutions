package main

// A repacker repacks trucks.
type repacker struct {
	bs boxStorage
	ps []palletStorage
}

func (r *repacker) addBox(b box) {
	r.bs.addBox(b)
}

func (r *repacker) addNewPallet() {
	r.ps = append(r.ps, palletStorage{false, pallet{}})
}

var packerRobot repacker

func organizePallet(in *truck) (out *truck) {
	out = &truck{id: in.id}

	//unloading tracks and moving boxes to the storage.
	for _, p := range in.pallets {
		for _, b := range p.boxes {
			packerRobot.addBox(b)
		}
	}

	//temp box value to do operations
	var currentBox box

	if len(packerRobot.ps) == 0 {
		packerRobot.addNewPallet()
	}

	palletN := 0

	for !packerRobot.bs.isEmpty() {
		for !packerRobot.ps[palletN].complete {
			biggestSpace := packerRobot.ps[palletN].findBiggestSpace()
			currentBox = packerRobot.bs.findBiggestBox(biggestSpace.width, biggestSpace.length)

			if currentBox.w > 0 && currentBox.l > 0 {
				//driver program uses different coordinate system
				currentBox.y = biggestSpace.x
				currentBox.x = biggestSpace.y
				packerRobot.ps[palletN].boxes = append(packerRobot.ps[palletN].boxes, currentBox)
			} else {
				//when no suitable box is found.
				palletN++
				packerRobot.addNewPallet()
				break
			}

			//when a pallet is full and ready to go.
			if packerRobot.ps[palletN].isFull() {
				out.pallets = append(out.pallets, pallet{packerRobot.ps[palletN].boxes})
				packerRobot.ps = packerRobot.ps[:palletN+copy(packerRobot.ps[palletN:], packerRobot.ps[palletN+1:])]
				packerRobot.addNewPallet()
			}
		}
	}

	//incomplete pallets are loaded to the last truck
	if out.id == idLastTruck {
		for _, p := range packerRobot.ps {
			if p.Items() > 0 {
				out.pallets = append(out.pallets, pallet{p.boxes})
			}
		}
	}

	return
}

func newRepacker(in <-chan *truck, out chan<- *truck) *repacker {
	go func() {
		for t := range in {
			// The last truck is indicated by its id. You might
			// need to do something special here to make sure you
			// send all the boxes.
			if t.id == idLastTruck {
			}

			t = organizePallet(t)
			out <- t
		}
		// The repacker must close channel out after it detects that
		// channel in is closed so that the driver program will finish
		// and print the stats.
		close(out)
	}()
	return &repacker{}
}
