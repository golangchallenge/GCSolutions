// Package drum is supposed to implement the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
package drum

import (
	"fmt"
	"strings"
)

const (
	magicLen       = 14 // Bytes up to and including the length (X) byte "SPLICE.......X"
	headerLen      = 50 // Bytes up to the first track header
	trackHeaderLen = 5  // Bytes per track header
	trackLen       = 16 // Number of notes per track
	barLen         = 4  // Number of notes per bar
)

// Header is the representation of the .splice file
// header information. (50 bytes)
type Header struct {
	Magic   [8]byte // Assumed 8 byte Magic number (http://en.wikipedia.org/wiki/Magic_number_%28programming%29#Magic_numbers_in_files)
	_       [5]byte // Padding until Len byte
	Len     byte    // Remaining bytes until the end of the valid data (1 byte)
	Version [32]byte
	Tempo   float32 // (4 bytes)
}

// Track is the representation of an indicidual
// .splice file drum track.
type Track struct {
	ID byte
	// (3 bytes padding in track header)
	NameLen byte
	Name    []byte
	Notes   []byte
}

// DecodeTrack decodes a byte slice into a track header
// and that track's notes. It returns a Track, and the number
// of bytes decoded.
func DecodeTrack(b []byte) (Track, int) {
	t := Track{}
	t.ID = b[0]
	// Skip 3 bytes padding
	t.NameLen = b[4]
	t.Name = b[trackHeaderLen : t.NameLen+trackHeaderLen]
	t.Notes = b[t.NameLen+trackHeaderLen : t.NameLen+trackHeaderLen+trackLen]
	return t, int(t.NameLen) + trackHeaderLen + trackLen
}

// Implement Stringer interface to output pattern data as text.
func (p *Pattern) String() string {
	// Strip NUL bytes from the version string
	version := strings.Trim(fmt.Sprintf("%s", p.Version), string(0x00))
	out := fmt.Sprintf("Saved with HW Version: %s\nTempo: %.3g\n", version, p.Tempo)
	for _, t := range p.Tracks {
		out += fmt.Sprintf("(%d) %s\t", t.ID, t.Name)
		for i, n := range t.Notes {
			if i%barLen == 0 { // If we're starting a new bar
				out += "|"
			}
			if n == 0 { // The note is off
				out += "-"
			} else { // The note is on
				out += "x"
			}
		}
		out += "|\n" // Close the track
	}
	return out
}
