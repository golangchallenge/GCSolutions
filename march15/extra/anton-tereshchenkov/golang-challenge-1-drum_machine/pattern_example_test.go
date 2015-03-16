package drum

import "fmt"

func Example_more_cowbell() {
	pattern, err := DecodeFile("fixtures/pattern_2.splice")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("pattern_2.splice")
	fmt.Println(pattern)

	track, err := pattern.FindTrack(5)
	if err != nil {
		fmt.Println(err)
		return
	}
	track.Play(0, 4, 6, 12, 14)

	fmt.Println("pattern_2-morebells.splice")
	fmt.Print(pattern)
	// Output:
	// pattern_2.splice
	// Saved with HW Version: 0.808-alpha
	// Tempo: 98.4
	// (0) kick	|x---|----|x---|----|
	// (1) snare	|----|x---|----|x---|
	// (3) hh-open	|--x-|--x-|x-x-|--x-|
	// (5) cowbell	|----|----|x---|----|
	//
	// pattern_2-morebells.splice
	// Saved with HW Version: 0.808-alpha
	// Tempo: 98.4
	// (0) kick	|x---|----|x---|----|
	// (1) snare	|----|x---|----|x---|
	// (3) hh-open	|--x-|--x-|x-x-|--x-|
	// (5) cowbell	|x---|x-x-|x---|x-x-|
}
