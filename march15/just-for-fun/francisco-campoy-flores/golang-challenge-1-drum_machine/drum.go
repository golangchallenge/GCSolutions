// Package drum provides decoding and printing functionality for .splice
// drum machine files.
//
// The package also provides a way of adding more cowbell to any pattern.
//
// The binary format for .splice drum machine files can be described with this
// grammar.
//
//	FILE = HEADER PATTERN
//
//	HEADER (14 bytes) =
//		[ 0 ..  5]	splice tag: "SPLICE"
//		[ 6 .. 13]	size of the following PATTERN in bytes: big endian uint64
//
//	PATTERN = PATTERN_HEADER TRACK*
//
//	PATTERN_HEADER (36 bytes) =
//		[ 0 .. 31]	HW version: nill terminated string
//		[32 .. 35]	Tempo: little endian float32
//
//	TRACK =
//		[ 0 .. 1]	track id: byte
//		[ 2 .. 5]	track name length: big endian int32 (n)
//		[ 6 .. x-1]	track name: string of the length decode previously
//		[ x .. x+15]	16 steps, 1 byte (0 or 1) each
//
// See http://golang-challenge.com/go-challenge1/ for more information
package drum

import (
	"bytes"
	"fmt"
)

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version string
	Tempo   float32
	Tracks  []*Track
}

// Track contains the information for a single track.
type Track struct {
	ID    uint8
	Name  string
	Steps [16]byte
}

// Pattern is a fmt.Stringer.
func (p Pattern) String() string {
	w := new(bytes.Buffer)
	fmt.Fprintf(w, "Saved with HW Version: %s\n", p.Version)
	fmt.Fprintf(w, "Tempo: %g\n", p.Tempo)
	for _, t := range p.Tracks {
		fmt.Fprintf(w, "(%d) %s\t", t.ID, t.Name)
		for i, b := range t.Steps {
			if i%4 == 0 {
				w.WriteByte('|')
			}
			if b == 0 {
				w.WriteByte('-')
			} else {
				w.WriteByte('x')
			}
		}
		w.WriteString("|\n")
	}
	return w.String()
}

// Because there's never enough cowbell.
func (p *Pattern) AddMoreCowbell() error {
	seen := make(map[uint8]bool)

	for _, t := range p.Tracks {
		seen[t.ID] = true
		if t.Name == "cowbell" {
			for i := range t.Steps {
				t.Steps[i] = 1
			}
			return nil
		}
	}

	// No cowbell??? What were you thinking?
	var id uint8
	for i := uint8(0); i <= 255; i++ {
		if !seen[i] {
			break
		}
	}
	if seen[id] {
		return fmt.Errorf("you have reached the maximum number of tracks")
	}
	p.Tracks = append(p.Tracks, &Track{
		ID:    id,
		Name:  "cowbell",
		Steps: [16]byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
	})
	return nil
}
