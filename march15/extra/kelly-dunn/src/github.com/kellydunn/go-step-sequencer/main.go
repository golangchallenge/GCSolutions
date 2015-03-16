package main

import (
	"flag"
	"fmt"
	drum "github.com/kellydunn/go-challenge-1"
	sequencer "github.com/kellydunn/go-step-sequencer/sequencer"
	"time"
)

// Entry point for go-step-sequencer
// Parses command line flags, which can be either:
//   - --pattern A filepath for the splice pattern on the filesystem.
//   - --kit A directory that contains all the samples for the tracks contained in the pattern.
func main() {
	var patternPath string
	var kitPath string

	flag.StringVar(
		&patternPath,
		"pattern",
		"patterns/pattern_1.splice",
		"-pattern=path/to/pattern.splice",
	)

	flag.StringVar(
		&kitPath,
		"kit",
		"kits",
		"-kit=path/to/kits",
	)

	flag.Parse()

	pattern, err := drum.DecodeFile(patternPath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	for _, track := range pattern.Tracks {
		filepath := kitPath + "/" + pattern.Version + "/" + track.Name + ".wav"

		track.Buffer, err = sequencer.LoadSample(filepath)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		track.Playhead = len(track.Buffer)

		fmt.Printf("loaded sample: %s\n", filepath)
	}

	fmt.Printf("%s\n", pattern)

	s, err := sequencer.NewSequencer()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	s.Pattern = pattern
	s.Timer.SetTempo(s.Pattern.Tempo)

	s.Start()

	for {
		time.Sleep(time.Second)
	}

}
