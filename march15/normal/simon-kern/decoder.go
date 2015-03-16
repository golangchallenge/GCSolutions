// Copyright 2015 by Simon Kern.
// Use of this source code is governed by a cc-style license.

// The author is aware of the fact that it might have been shorter and / or easier
// to just decode strings into an array of (type) byte and convert it during the
// invocation of the Payload's and Track's String() method.
//
// The use of the blank identifier within the structs to dump some bytes, that are not necessary in
// order to generate the output, could shortened the code as well.
//
// However, it was important to the author that the struct does not contain anything that
// most likely has not been part of the original drum machine.

package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
)

const (
	vStrLen       = 32 //Length of versionString in bytes
	nSteps        = 16 // number of Steps per Track
	identifierLen = 6
	headerSize    = identifierLen + 8 // headerSize in bytes - 8 because payloadSize is an int64
)

var (
	spliceIdentifier = [identifierLen]byte{'S', 'P', 'L', 'I', 'C', 'E'}
)

type spliceDecoder struct {
	r   io.Reader
	p   *Pattern
	err error
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Header  Header
	Payload Payload
}

// Header contains those information that are contained in a .splice file
// that are not part of the Payload.
type Header struct {
	Identifier  [identifierLen]byte
	PayloadSize int64 //Size of Payload in bytes (excluding Header)
}

// Payload contains the actual Payload contained in a .splice file.
type Payload struct {
	Version string
	Tempo   float32
	Tracks  []Track
}

// Track is part of the Payload of a .splice file.
type Track struct {
	ID    int32
	Name  string
	Steps []int8
}

// creates and returns a new spliceDecoder
func newSpliceDecoder(r io.Reader) *spliceDecoder {
	p := new(Pattern)
	return &spliceDecoder{r, p, nil}
}

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	fd, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	dec := newSpliceDecoder(fd)

	dec.DecodeHeader()

	dec.CheckFileSize(fd)

	dec.DecodePayload()
	if dec.err != nil && dec.err != io.EOF {
		return nil, dec.err
	}

	return dec.p, nil
}

// DecodeHeader decodes the Header part of the input file
func (dec *spliceDecoder) DecodeHeader() {
	// If SpliceDecoder already encountered an error, return
	if dec.err != nil {
		return
	}

	dec.err = binary.Read(dec.r, binary.LittleEndian, &dec.p.Header.Identifier)
	if dec.err != nil {
		return
	}

	if dec.p.Header.Identifier != spliceIdentifier {
		dec.err = errors.New("Invalid input file.")
		return
	}

	dec.err = binary.Read(dec.r, binary.BigEndian, &dec.p.Header.PayloadSize)
	if dec.err != nil {
		return
	}
}

// CheckFileSize checks whether the actual size of a .splice file (fd) matches
// the size of the file's Header + the size of its Payload (as stated in its Header).
//
// If the file is larger it replaces the spliceDecoder's reader with a LimitReader
// that prevents spliceDecoder from reading beyond the Payload.
//
// If the file is smaller than it should be, CheckFileSize sets dec.err, which will cause
// DecodeFile() to fail.
func (dec *spliceDecoder) CheckFileSize(fd *os.File) {
	// If SpliceDecoder already encountered an error, return
	if dec.err != nil {
		return
	}

	//Get fd's fileSize
	fileInfo, err := fd.Stat()
	if err != nil {
		dec.err = err
		return
	}
	fileSize := fileInfo.Size()

	// Calculate expected FileSize based PayloadSize (as stated in its Header)
	// and the actual size of its header Header size
	expectedSize := dec.p.Header.PayloadSize + headerSize
	if fileSize > expectedSize {

		// This Limitreader prevents spliceDecoder from reading data that is behind the
		// payload (possibly garbage)
		payloadReader := io.LimitReader(dec.r, dec.p.Header.PayloadSize)
		dec.r = payloadReader

	} else if fileSize < expectedSize {

		// If fileSize is smaller than expectedSize, we are missing payload and therefore fail.
		dec.err = fmt.Errorf("Payload is not complete. Expected: %d bytes - Got: %d bytes", expectedSize, fileSize)

	}
}

// DecodePayload decodes the Payload that is contained in a .splice file.
func (dec *spliceDecoder) DecodePayload() {
	// If SpliceDecoder already encountered an error, return
	if dec.err != nil {
		return
	}

	// Decode versionStr
	versionStr, err := DecodeBinaryString(dec.r, vStrLen)
	if err != nil {
		dec.err = err
		return
	}
	dec.p.Payload.Version = versionStr

	//Decode Tempo, which is a float32
	dec.err = binary.Read(dec.r, binary.LittleEndian, &dec.p.Payload.Tempo)
	if dec.err != nil {
		return
	}

	dec.DecodeTracks()
}

// DecodeTracks decodes Tracks until it encounters an error
// If the error encountered is EOF the execution of DecodeTracks can be considered successful.
func (dec *spliceDecoder) DecodeTracks() {
	// If SpliceDecoder already encountered an error, return
	if dec.err != nil {
		return
	}

	// Decode tracks as long as we do not encounter an error
	// If the error encountered is EOF DecodeFile() will exit gracefully
	for dec.err == nil {
		// create new Track
		t := new(Track)

		// Decode Track.ID
		dec.err = binary.Read(dec.r, binary.LittleEndian, &t.ID)
		if dec.err != nil {
			return
		}

		// Decode nameLen, which is the size of t.Name in bytes
		// nameLen itself will not be stored in our Track struct.
		var nameLen uint8
		dec.err = binary.Read(dec.r, binary.LittleEndian, &nameLen)
		if dec.err != nil {
			return
		}

		// Decode nameLen bytes to trackName
		trackName, err := DecodeBinaryString(dec.r, int(nameLen))
		if err != nil {
			dec.err = err
			return
		}
		t.Name = trackName

		// Decode steps
		steps := make([]int8, nSteps)
		err = binary.Read(dec.r, binary.LittleEndian, &steps)
		if err != nil {
			dec.err = err
			return
		}
		t.Steps = steps

		// append track to our Payload's tracks
		dec.p.Payload.Tracks = append(dec.p.Payload.Tracks, *t)
	}

}

// DecodeBinaryString reads n bytes from r and extracts the string from those bytes.
// It therefore determines the first occurence of a nullbyte, which terminates the string.
func DecodeBinaryString(r io.Reader, n int) (string, error) {
	// create a buffer for n bytes
	buf := make([]byte, n)

	// Read n bytes from r and store them in the buffer
	err := binary.Read(r, binary.LittleEndian, buf)
	if err != nil {
		return "", err
	}
	//find nullbyte in the byteSequence, so that we can determine the end of the string
	endOfStr := bytes.IndexByte(buf, 0)

	// If there is no Nullbyte just terminate at the end of buf.
	if endOfStr == -1 {
		endOfStr = len(buf)
	}

	return string(buf[:endOfStr]), nil
}
