// Package drum implements the decoding of .splice drum machine files.
//
// The file format consits of a header and then a number of tracks
// Header
//  6 bytes magic "SPLICE"
//  8 bytes big endian int64 length of the rest of the file
// 32 bytes version string, paded with 0x00
//  4 bytes little endian float32 tempo/bpm
//
// reducing the read length by 32 and 4 leaves us with track data
//
// Track
//  1 byte track id
//  4 bytes big endian int32 name length
//  x bytes name
// 16 bytes 0x00 or 0x01 for each step
//
// Anything past the specified length in the header is ignored
//
// See golang-challenge.com/go-challenge1/ for more information
package drum

import (
	"bytes"
	"fmt"
)

var spliceMagic = []byte{'S', 'P', 'L', 'I', 'C', 'E'}

// AddTrack adds or replaces a track with the same id
func (p *Pattern) AddTrack(t Track) {
	found := false
	for k, v := range p.Tracks {
		if v.ID == t.ID {
			p.Tracks[k] = t
			found = true
			break
		}
	}
	if !found {
		p.Tracks = append(p.Tracks, t)
	}
}

// String returns a visual representation of the splice file
func (p *Pattern) String() string {
	buffer := &bytes.Buffer{}

	buffer.WriteString("Saved with HW Version: ")
	buffer.WriteString(p.Version)
	buffer.WriteString("\nTempo: ")
	buffer.WriteString(fmt.Sprintf("%g\n", p.Tempo))
	for _, t := range p.Tracks {
		buffer.WriteString(fmt.Sprintf("(%d) ", t.ID))
		buffer.WriteString(t.Name)
		buffer.WriteString("\t")
		// 16
		for k, s := range t.Steps {
			if k%4 == 0 {
				buffer.WriteByte('|')
			}
			if s {
				buffer.WriteByte('x')
			} else {
				buffer.WriteByte('-')
			}
		}
		buffer.WriteString("|\n")
	}
	return buffer.String()
}
