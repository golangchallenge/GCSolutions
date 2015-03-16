package main

import (
	"fmt"
	"github.com/rosssalas/drum"
)

func main() {
	// let's load each splice file and decode it
	// also try to load sixth file to see error
	for i := 1; i < 7; i++ {
		fn := fmt.Sprintf("fixtures/pattern_%d.splice", i)
		p, err := drum.DecodeFile(fn)
		if err != nil {
			fmt.Println("error processing file:", fn)
		} else {
			fmt.Println(fn, p)
		}
	}

	// let's work with one splice file
	fn := fmt.Sprintf("fixtures/pattern_1.splice")
	p, err := drum.DecodeFile(fn)
	if err != nil {
		fmt.Println("error processing file:", fn)
		return
	}

	//  putting it all together, a simple player
	//  loop through notes 0 - 15 and play track if there is 1 in step
	fmt.Println(p)
	fmt.Println("step  track\n*************************")
	for step := 0; step < 16; step++ {
		fmt.Printf("[%02d]: ", step)
		// check if any track should be played
		for _, track := range p.Tracks {
			if track.PlayNote(step) {
				fmt.Printf("[%s] ", track.Name)
				//Instrument[track.Id].Start()
			}
		}
		fmt.Println()
		//delay based on tempo
	}
}
