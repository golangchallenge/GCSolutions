/*
Package drum implements the decoding of .splice drum machine files.
See golang-challenge.com/go-challenge1/ for more information
This file contains implemntation details of the decoder.
*/
package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
)

// binaryHeader is a struct used to read the header part of the file
type binaryHeader struct {
	Splice    [6]byte  // First 6 bytes are always "SPLICE"
	_         [7]byte  // 7 bytes padding
	Length    uint8    // The length in bytes of the rest of the file
	HwVersion [32]byte // HW version string + padding bytes up 32
	Tempo     float32  // 32 bit float tempo
}

// headerLength is the length of a binaryHeader in bytes
const headerLength = 36

// binaryTrackHeader is a struct used to read the first part of a track
type binaryTrackHeader struct {
	Id         uint32 // Instrument id number
	NameLength uint8  // The length in bytes of the instrument name following the header
}

// trackHeaderLength is the length of a binaryTrackHeader in bytes
const trackHeaderLength = 5

// readHeader reads the header from "file" and put the data into a Header structure.
// "remaining" is the count in bytes that should be read after the header.
func readHeader(file io.Reader) (header *Header, remaining uint8, err error) {
	var rawHead binaryHeader
	err = binary.Read(file, binary.LittleEndian, &rawHead)
	if err != nil {
		return nil, 0, err
	}

	if string(rawHead.Splice[:]) != "SPLICE" {
		return nil, 0, InvalidFile
	}

	// Convert rawHead.HwVersion to a UTF-8 string, and then trim off any nulls.
	version := string(bytes.Runes(rawHead.HwVersion[:]))
	version = strings.Trim(version, "\000")

	head := Header{HardwareVersion: version, Tempo: rawHead.Tempo}
	return &head, rawHead.Length - headerLength, nil
}

// readTrack reads a single track from "file" and put the data into a Track structure.
// "remaining" should be the count in bytes left to read from "file".
func readTrack(file io.Reader, remaining *uint8) (*Track, error) {
	// Read the first part of the track head, id + name length.
	*remaining -= trackHeaderLength
	if *remaining < 0 {
		return nil, FileTooShort
	}
	var trackHead binaryTrackHeader
	err := binary.Read(file, binary.LittleEndian, &trackHead)
	if err != nil {
		return nil, err
	}

	// Read the string based on the length read above.
	*remaining -= trackHead.NameLength
	if *remaining < 0 {
		return nil, FileTooShort
	}
	name := make([]byte, trackHead.NameLength)
	err = binary.Read(file, binary.LittleEndian, name)
	if err != nil {
		return nil, err
	}

	// Read the actual track data.
	*remaining -= TrackDataLength
	if *remaining < 0 {
		return nil, FileTooShort
	}
	var data TrackData
	err = binary.Read(file, binary.LittleEndian, &data)
	if err != nil {
		return nil, err
	}

	// Build and return the finished track structure.
	track := Track{Id: trackHead.Id, Name: string(bytes.Runes(name)), Data: data}
	return &track, nil
}

// String returns a string representation of the given pattern.
// Written to conform with the style expected by the challenge tests.
func (p *Pattern) String() string {
	// 2 lines for the version and tempo
	str := fmt.Sprintf("Saved with HW Version: %v\n", p.Header.HardwareVersion)
	str = str + fmt.Sprintf("Tempo: %v\n", p.Header.Tempo)
	// Add a line for each track.
	for _, track := range p.Tracks {
		// Build the |--x-|----|----|---x| part
		track_str := "|"
		for j := 0; j < 4; j++ {
			for k := 0; k < 4; k++ {
				if track.Data[4*j+k] == 0 {
					track_str += "-"
				} else {
					track_str += "x"
				}
			}
			track_str += "|"
		}
		str = str + fmt.Sprintf(`(%v) %v	%v%v`, track.Id, track.Name, track_str, "\n")
	}
	return str
}
