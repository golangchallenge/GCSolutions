// Package drum is supposed to implement the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
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
	headerSize = 14
)

var spliceMagic = []byte{0x53, 0x50, 0x4c, 0x49, 0x43, 0x45}

// Open a Splice file and interpret the data into a Pattern/Tracks
func Open(path string) (*Pattern, error) {
	fi, err := os.Stat(path)

	// TODO better error handling
	if err != nil {
		return nil, err
	}

	if fi.Size() < headerSize {
		return nil, fmt.Errorf("Splice file must be at least 14 bytes (input file is %d bytes)", fi.Size())
	}

	// Read first 14 bytes to check magic header and required minimum filesize
	spliceFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer spliceFile.Close()

	valid, err := checkSpliceMagic(spliceFile)
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, errors.New("File in not a valid splce file, magic header incorrect")
	}

	// if packed on 1-byte alignment, dataSize could be a int64 in big-endian
	// hard to tell from example data and not sure why you'd mix big/little endian in the same file
	// but a single byte value seems too limiting regarding number of tracks
	var dataSize int64
	binary.Read(spliceFile, binary.BigEndian, &dataSize)
	minimumFileSize := headerSize + dataSize

	// By obeying the data size and not the file size we support fixture 5 with extra unused data
	if minimumFileSize > fi.Size() {
		return nil, fmt.Errorf("File is truncated, required file size is %d - actual file size is %d", minimumFileSize, fi.Size())
	}

	// More header data

	// 32 byte character ASCII string
	version, err := readFixedNullTermString(spliceFile, 32)
	if err != nil {
		return nil, err
	}

	//  4 byte little endian floating point
	var tempo float32
	err = binary.Read(spliceFile, binary.LittleEndian, &tempo)
	if err != nil {
		return nil, err
	}

	var tracks []*Track

	for {
		track := readTrack(spliceFile)
		tracks = append(tracks, track)

		pos, _ := spliceFile.Seek(0, os.SEEK_CUR)
		if pos-headerSize >= dataSize {
			break
		}

	}

	return &Pattern{
		Version: version,
		Tempo:   tempo,
		Tracks:  tracks,
	}, nil

}

func checkSpliceMagic(spliceFile io.Reader) (bool, error) {
	spliceHeader := make([]byte, len(spliceMagic))
	n, err := spliceFile.Read(spliceHeader)
	if err != nil && err != io.EOF {
		return false, err
	}
	if n != len(spliceMagic) {
		return false, errors.New("Could not read magic header, insufficient data")
	}

	// Verify header

	// first 6 bytes are SPLICE
	if bytes.Compare(spliceHeader, spliceMagic) != 0 {
		return false, errors.New("File is not a valid splice file, magic header incorrect")
	}

	return true, nil
}
