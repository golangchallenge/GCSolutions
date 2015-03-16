package main

import (
	"flag"
	"fmt"
	"strings"

	drum "bitbucket.org/stridatum/go-challenge-1"
)

var (
	addCowbell = flag.Bool("add", false, "add in some cowbell if not present")

	augCowbell = flag.Bool("augment", false, "augment every note with cowbell")
	antCowbell = flag.Bool("anti", false, "fill all silence with cowbell")
	allCowbell = flag.Bool("all", false, "all cowbell all the time")
)

func main() {
	flag.Parse()

	dm, err := drum.DecodeFile(flag.Arg(0))
	if err != nil {
		fmt.Printf("decoding failed - %v\n", err)
		return
	}

	var otherPartSteps uint16
	var maxID int32
	cb := -1
	for i, p := range dm.Parts {
		if p.Name == "cowbell" {
			cb = i
		} else {
			otherPartSteps |= p.Steps
		}
		if p.ID > maxID {
			maxID = p.ID
		}
	}
	if cb == -1 {
		if *addCowbell {
			cb = len(dm.Parts)
			dm.Parts = append(dm.Parts, &drum.Part{ID: maxID + 1, Name: "cowbell"})
		} else {
			fmt.Println("no cowbell found :'( -- you can add some with -add")
			return
		}
	}

	if *augCowbell {
		dm.Parts[cb].Steps |= otherPartSteps
	}
	if *antCowbell {
		dm.Parts[cb].Steps = 0xFFFF &^ otherPartSteps
	}
	if *allCowbell {
		dm.Parts[cb].Steps = 0xFFFF
	}

	fmt.Print(dm)

	newFilename := strings.Replace(flag.Arg(0), ".splice", "-morebells.splice", -1)
	err = dm.EncodeFile(newFilename)
	if err != nil {
		fmt.Printf("encoding failed - %v\n", err)
		return
	}
	fmt.Println("Wrote the new cowbells to:", newFilename)
}
