package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	// Splice version string
	Version string
	// Tempo of the pattern
	Tempo float32
	// Instruments
	Instruments []Instrument
}

// Instrument is the high level represestation of a
// single instrument pattern contained in a .splice
// file.
type Instrument struct {
	// Instrument number
	ID uint8
	// Instrument name
	Name string
	// Instrument 16 beat measure
	Measure Measure
}

// Measure is the 16 beat measure of a .splice file.
type Measure [16]bool

type rawSplice struct {
	// SPLICE with four padding bytes
	header [10]byte
	// A 32 bit int containing the file length
	length uint32
	// version string
	version [32]byte
	// floating point LSB -> MSB
	tempo float32
	// the rest of the file are rawMeasures
	measures []rawMeasure
}

type rawMeasure struct {
	// each measure has a number associated with it
	id uint8
	// this is similar to a FORTH string len|string
	// nameLength is potentially a 16 bit int.
	// in the samples however, it's never above 30
	nameLength uint8
	name       []byte
	// This is the actualy measure with each byte representing a
	// beat.
	measure [16]byte
}

// Pattern converts a rawSplice to a high level Pattern
func (r rawSplice) Pattern() *Pattern {
	p := &Pattern{}
	// strip out nul bytes
	p.Version = string(bytes.Trim(r.version[0:], "\000"))

	p.Tempo = r.tempo

	p.Instruments = make([]Instrument, 0)
	// convert rawMeasure to an Instrument
	for _, r := range r.measures {
		var m Measure
		for i := range m {
			m[i] = r.measure[i] == 1
		}

		p.Instruments = append(p.Instruments, Instrument{
			ID:      r.id,
			Name:    string(r.name[0:]),
			Measure: m,
		})
	}
	return p
}

func readRawSliceInstrucment(file *os.File) (*rawMeasure, error) {
	p := &rawMeasure{}
	// Get the 1 byte measure number
	if err := binary.Read(file, binary.LittleEndian, &p.id); err != nil {
		return nil, err
	}

	// this could be part of a 32 bit int, but it would be BigEndian, and that
	// doesn't seem to be the case in the examples
	padding := make([]byte, 3)
	if err := binary.Read(file, binary.LittleEndian, &padding); err != nil {
		return nil, err
	}

	// Get the 8 bit int measure name length
	if err := binary.Read(file, binary.LittleEndian, &p.nameLength); err != nil {
		return nil, err
	}

	// Get the measure name
	p.name = make([]byte, p.nameLength)
	if err := binary.Read(file, binary.LittleEndian, p.name); err != nil {
		return nil, err
	}

	// Get the 16 bytes measure
	if err := binary.Read(file, binary.LittleEndian, &p.measure); err != nil {
		return nil, err
	}

	return p, nil
}

func readRawSliceInstrucments(file *os.File, length uint32) ([]rawMeasure, error) {
	r := []rawMeasure{}
	for {
		// read a raw instrument from the splice
		p, err := readRawSliceInstrucment(file)
		// if we're done reading based on the provided file length or we reach
		// the end of the file, we exit normally
		if length == 0 || err == io.EOF {
			return r, nil
		} else if err != nil {
			// Read error
			return nil, err
		}

		r = append(r, *p)

		// decrement how much we've read from the file
		// id(1) + nameLength(4) + measure(16) = 21
		length -= uint32(len(p.name) + 21)
	}
}

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed measure which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var r rawSplice
	// Read header that contains SPLICE\0 with 4 padding bytes
	if err := binary.Read(file, binary.LittleEndian, &r.header); err != nil {
		return nil, err
	}

	// Read remaining length of file
	if err := binary.Read(file, binary.BigEndian, &r.length); err != nil {
		return nil, err
	}
	length := r.length

	// Read version string
	if err := binary.Read(file, binary.LittleEndian, &r.version); err != nil {
		return nil, err
	}
	length -= uint32(len(r.version))

	// Read tempo
	if err := binary.Read(file, binary.LittleEndian, &r.tempo); err != nil {
		return nil, err
	}
	length -= 4

	// Read instruments measures
	if r.measures, err = readRawSliceInstrucments(file, length); err != nil {
		return nil, err
	}

	return r.Pattern(), nil
}

func (m Measure) String() string {
	var result string
	for i, v := range m {
		if i%4 == 0 {
			result += "|"
		}
		if v {
			result += "x"
		} else {
			result += "-"
		}
	}
	result += "|"

	return result
}

func (p Pattern) String() string {
	result := fmt.Sprintf("Saved with HW Version: %s\nTempo: %v\n", p.Version, p.Tempo)
	for _, i := range p.Instruments {
		result += fmt.Sprintf("(%d) %s\t%s\n", i.ID, i.Name, i.Measure.String())
	}
	return result
}
