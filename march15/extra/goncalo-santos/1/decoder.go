package drum

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	// Open file and get info
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not open file %s", path)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, errors.New("could not get file information")
	}

	pattern := &Pattern{}
	if err := pattern.decodeHeader(file); err != nil {
		return nil, err
	}

	// Check if the size in the header isn't bigger than the real size of the file
	// pattern.header.FileSize+14 -> Total file size
	if pattern.header.FileSize+14 > fileInfo.Size() {
		return nil, errors.New("this file is corrupted")
	}

	if err := pattern.decodeInstruments(file); err != nil {
		return nil, err
	}

	return pattern, nil
}

//Read header
func (p *Pattern) decodeHeader(r io.Reader) error {
	if err := binary.Read(r, binary.BigEndian, &p.header.SpliceTag); err != nil {
		return errors.New("couldn't read the SPLICE tag")
	}

	if err := binary.Read(r, binary.BigEndian, &p.header.FileSize); err != nil {
		return errors.New("couldn't read the file size")
	}

	if err := binary.Read(r, binary.BigEndian, &p.header.VersionBuf); err != nil {
		return errors.New("couldn't read the version")
	}

	if err := binary.Read(r, binary.LittleEndian, &p.header.BPM); err != nil {
		return errors.New("couldn't read the BPM")
	}

	// Check if it has SPLICE in the beginning
	if string(p.header.SpliceTag[:]) != "SPLICE" {
		return errors.New("invalid file, no SPLICE in the beginning")
	}

	return nil
}

// Read instruments
func (p *Pattern) decodeInstruments(r io.Reader) error {
	offset := int64(0)
	p.instruments = make(map[byte]*instrumentStruct)

	// p.header.FileSize-36 -> Size of the instruments part of the file
	// 36 = sizeof(VersionBuf)+sizeof(BPM)
	for offset < p.header.FileSize-36 {
		var instrument instrumentStruct
		if err := binary.Read(r, binary.BigEndian, &instrument.ID); err != nil {
			return errors.New("could not read the instrument's ID")
		}

		// This is for printing the instruments in the same order as they were read
		p.instrumentsOrder = append(p.instrumentsOrder, instrument.ID)

		if err := binary.Read(r, binary.BigEndian, &instrument.NameLength); err != nil {
			return errors.New("could not read the length of the instrument's name")
		}

		name := make([]byte, instrument.NameLength)
		if err := binary.Read(r, binary.BigEndian, &name); err != nil {
			return errors.New("could not read the instrument's name")
		}

		instrument.Name = string(name)

		if err := binary.Read(r, binary.BigEndian, &instrument.Pattern); err != nil {
			return errors.New("could not read the instrument's pattern")
		}

		p.instruments[instrument.ID] = &instrument

		offset += 1 + 4 + int64(instrument.NameLength) + 16
	}

	if offset != p.header.FileSize-36 {
		return errors.New("this file is corrupted")
	}

	return nil
}
