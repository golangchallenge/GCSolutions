package drum_test

import (
	"bytes"
	"fmt"

	"github.com/rogpeppe/misc/drum"
)

func Example_cowbell() {
	// This example demonstrates how to add more cowbell to
	// a drum machine pattern.

	p, err := drum.Decode(bytes.NewReader(cowbellExampleData))
	if err != nil {
		panic(err)
	}
	var cowbellTrack *drum.Track
	maxChan := 0
	for i := range p.Tracks {
		tr := &p.Tracks[i]
		if tr.Name == "cowbell" {
			cowbellTrack = tr
			break
		}
		if tr.Channel > maxChan {
			maxChan = tr.Channel
		}
	}
	if cowbellTrack == nil {
		p.Tracks = append(p.Tracks, drum.Track{
			Channel: maxChan + 1,
			Name:    "cowbell",
		})
		cowbellTrack = &p.Tracks[len(p.Tracks)-1]
	}
	for i := range cowbellTrack.Beats {
		if i%2 == 0 {
			cowbellTrack.Beats[i] = true
		}
	}
	fmt.Print(p)

	data, err := p.MarshalBinary()
	if err != nil {
		panic(err)
	}

	// To save the file, write data to a new file.
	_ = data
	// Output: Saved with HW Version: 0.808-alpha
	//Tempo: 120
	//(0) kick	|x---|x---|x---|x---|
	//(1) snare	|----|x---|----|x---|
	//(2) clap	|----|x-x-|----|----|
	//(3) hh-open	|--x-|--x-|x-x-|--x-|
	//(4) hh-close	|x---|x---|----|x--x|
	//(5) cowbell	|x-x-|x-x-|x-x-|x-x-|
}
