// Package drum is supposed to implement the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
package drum

// All the decode functions here assume that the data they are trying to decode
// is always at the front of the reader that is passed to the functions

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// fill slice p with data from file until p is full
// return nil if there was no error
func readSlice(file io.Reader, p []byte) error {
	// loop until either p is filled or an error occurs
	for {
		n, err := file.Read(p)

		// shrink p down the the amount left to read
		p = p[n:]

		if err != nil {
			if err == io.EOF && len(p) == 0 {
				// end of file reached after slice has finished reading
				// not an issue
				return nil
			}
			return err
		}

		if len(p) == 0 {
			// p is has been filled
			return nil
		}
	}
}

// Check that the first 6 characters should be "SPLICE" (without quotation marks)
func decodeFormatString(file io.Reader) error {
	p := make([]byte, 6)

	err := readSlice(file, p)
	if err != nil {
		return err
	}

	reqVal := []byte{'S', 'P', 'L', 'I', 'C', 'E'}

	if bytes.Compare(p, reqVal) != 0 {
		return fmt.Errorf("File must begin with \"SPLICE\" identifier string")
	}

	return nil
}

// Decode legth of remainder of (usable) file
// This is just an assumption, but
// this value is a 64-bit, BigEndian int
//  ( It is possible that some of those bytes are for the format string,
//  ( but the format string doesn't need them, so this seems like a fairly safe choice
func decodeFileLength(file *os.File) (endPos int64, err error) {

	var remLength int64
	err = binary.Read(file, binary.BigEndian, &remLength)
	if err != nil {
		return 0, err
	}

	startPos, err := file.Seek(0, 1) // get current offset into file
	if err != nil {
		return 0, err
	}

	return startPos + remLength, nil
}

// Decode the HWVersion HWVersion
// takes up 32 bytes in the file
func decodeHWVersion(file io.Reader) (version string, err error) {
	// the first 32 bytes are a string which is the HW version
	// ex. "0.808-alpha" followed by zeros

	p := make([]byte, 32)
	err = readSlice(file, p)
	if err != nil {
		return "", err
	}

	// trim null characters off of p
	for i := range p {
		if p[i] == 0 {
			// ensure that all following characters are also null
			for _, c := range p[i:] {
				if c != 0 {
					return "", fmt.Errorf(
						"Non-null character after null in HWVersion string: %#x", c,
					)
				}
			}
			// actually trim off null characters
			p = p[:i]
			break
		}
	}

	version = string(p)
	return version, nil
}

// Decode tempo
// tempo is stored in a float32, little-endian
func decodeTempo(file io.Reader) (tempo float32, err error) {
	err = binary.Read(file, binary.LittleEndian, &tempo)
	if err != nil {
		return 0, err
	}

	return tempo, nil
}

// Decode 1 track
func decodeTrack(file io.Reader) (track Track, err error) {
	//first is the track's id as a 32-bit uint, little-endian
	err = binary.Read(file, binary.LittleEndian, &track.ID)
	if err != nil {
		return track, err
	}

	// then is the length of the name string (1 byte)
	var strLen byte
	err = binary.Read(file, binary.LittleEndian, &strLen)
	if err != nil {
		return track, err
	}

	// read the name string
	name := make([]byte, strLen)
	err = readSlice(file, name)
	if err != nil {
		return track, err
	}
	track.Name = string(name)

	// read the 16 bytes (16 notes) for track.Beats
	err = readSlice(file, track.Beats[:])
	if err != nil {
		return track, err
	}

	return track, nil
}
