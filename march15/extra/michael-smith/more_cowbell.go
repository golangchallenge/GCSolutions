package main

import (
	"drum"
	"fmt"
	"os"
)

// Parses a .splice file, adds more cowbell, and then writes it back out.
//
// Usage:
//   more_cowbell [in_file] [out_file]
func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s [in_file] [out_file]\n", os.Args[0])
		os.Exit(1)
	}

	inPath := os.Args[1]
	outPath := os.Args[2]

	// decode from file
	p, err := drum.DecodeFile(inPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	trackName := "cowbell"
	cowbellTrack := p.FindTrack(trackName)
	var steps drum.Steps
	for i := range steps {
		steps[i] = true
	}

	// more cowbell !!!1
	if cowbellTrack != nil {
		cowbellTrack.Steps = steps
	} else {
		cowbellTrack = &drum.Track{
			Name:  trackName,
			Steps: steps,
		}
		p.AddTrack(*cowbellTrack)
	}
	fmt.Println(p)

	// write new file
	if err := drum.EncodeFile(outPath, p); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
