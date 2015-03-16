// Package drum is supposed to implement the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
package drum

import (
	"bytes"
	"fmt"
)

const (
	TRACK_PADDING  = 3  // padding size used inside track data
	HEADER_PADDING = 12 // initial header padding
)

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Header HeaderInfo
	Tracks []Track
}

// Tempo as a type which helps to impletement custom string function.
type tempo float32

// HeaderInfo contains header information of the .splice file.
type HeaderInfo struct {
	Size    uint16
	Version [32]byte
	Tempo   tempo
}

// Track represent a single track in a drum file.
type Track struct {
	ID    uint8
	Name  []byte
	Steps [16]byte
}

// String returns the printable text of tempo.
func (t tempo) String() string {
	// check if there is any precision
	if float32(t) == float32(int32(t)) {
		return fmt.Sprintf("Tempo: %d", int(t))
	}
	return fmt.Sprintf("Tempo: %.1f", t)
}

// String returns the printable text of HeaderInfo.
func (h HeaderInfo) String() string {
	arraySize := bytes.Index(h.Version[:], []byte{0})
	output := fmt.Sprintf("Saved with HW Version: %s\n", h.Version[:arraySize])
	output += fmt.Sprintln(h.Tempo)
	return output
}

// String returns the printable text of Pattern.
func (p Pattern) String() string {
	output := fmt.Sprint(p.Header)
	for _, t := range p.Tracks {
		output += fmt.Sprintln(t)
	}
	return output
}

// String returns the printable text of Track.
func (t Track) String() string {
	output := fmt.Sprintf("(%d) %s\t", t.ID, t.Name[:])
	for index, step := range t.Steps {
		if index%4 == 0 {
			output += fmt.Sprint("|")
		}
		if step == 0 {
			output += fmt.Sprint("-")
		} else {
			output += fmt.Sprint("x")
		}
	}
	output += fmt.Sprint("|")
	return output
}
