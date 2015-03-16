// Package drum is supposed to implement the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io/ioutil"
	"math"
)

// ParseSplice initializes the parsing of a drum machine file. The track data
// always begins at the same offset (50).
//
// If the read of a drum machine file fails or the header is invalid,
// the decoding will be stopped immediately.
func ParseSplice(path string, p *Pattern) error {
	offset := 50
	splice, err1 := ioutil.ReadFile(path)
	err2 := checkHeader(&splice)

	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}

	// Fetch the version from a fixed data range.
	p.version = string(bytes.Trim(splice[14:33], "\x00"))
	fetchTempo(splice, p)
	fetchTracks(splice, &offset, p)

	return nil
}

// checkHeader analyzes the vailidity of a drum machine's header file.
//
// If the SPLICE keyword appears twice, only the data before
// the second appearance of the keyword is used for decoding.
//
// An error is returned if the SPLICE keyword appears only once and hasn't been
// at the very beginning of the file.
func checkHeader(splice *[]byte) error {
	header := []byte("SPLICE")

	if bytes.Count(*splice, header) > 1 {
		if index := bytes.LastIndex(*splice, header); index > 0 {
			*splice = (*splice)[0:index]
		}
	} else {
		if bytes.Index(*splice, header) != 0 {
			return errors.New("no valid header found")
		}
	}
	return nil
}

// fetchTempo fetches the tempo of the drum machine file from a fixed position.
func fetchTempo(splice []byte, p *Pattern) {
	b := binary.LittleEndian.Uint32(splice[46:51])
	f := math.Float32frombits(b)
	p.tempo = f
}

// fetchTracks fetches one track starting from the specified offset. It will
// repeat fetching tracks until the end of the drum machine data is reached.
//
// If there is not enough data for a valid track left, stop to fetch tracks.
func fetchTracks(splice []byte, offset *int, p *Pattern) {
	if len(splice) <= *offset+22 {
		return
	}

	track := Track{}
	track.id = int(splice[*offset])
	*offset = *offset + 5

	// Finds the name of the track by iterating over splice data.
	//
	// If a zero or one appears, stop iterating and continue with the steps.
	for {
		if IsByteZero(splice[*offset]) || IsByteOne(splice[*offset]) {
			break
		} else {
			track.name += string(splice[*offset])
			*offset++
		}
	}

	// Finds the steps of the track by iterating over the following 16 bytes
	// after the track's id and name.
	//
	// If a step is not zero or one, don't consider it as a valid step.
	for i := 0; i < 16; i++ {
		if IsByteZero(splice[*offset]) || IsByteOne(splice[*offset]) {
			track.steps[i] = splice[*offset]
		}
		*offset++
	}

	p.tracks = append(p.tracks, track)

	fetchTracks(splice, offset, p)
}

// IsByteZero checks if the passed byte equals to the number 0.
func IsByteZero(data byte) bool {
	return bytes.Equal([]byte{data}, []byte{0x00})
}

// IsByteOne checks if the passed byte equals to the number 1.
func IsByteOne(data byte) bool {
	return bytes.Equal([]byte{data}, []byte{0x01})
}
