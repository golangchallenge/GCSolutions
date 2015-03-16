// Package drum is supposed to implement the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
package drum

import (
	"bytes"
	"fmt"
)

// Representation of the header
/*
HEADER FORMAT:
	6 bytes:	"SPLICE"
	8 bytes:	File size (since version's beginning) -> int64 (BigEndian)
	32 bytes:	Version (literal)
	4 bytes:	BPM -> float32 (LittleEndian)

TOTAL = 50 bytes
*/
type headerStruct struct {
	SpliceTag  [6]byte
	FileSize   int64
	VersionBuf [32]byte
	BPM        float32
}

// Representation of each instrument in a file
/*
INSTRUMENT FORMAT:
	1 byte:		Instrument's ID -> byte
	4 bytes:	Name length -> int32 (BigEndian)
	X bytes:	Name
	16 bytes:	Pattern
*/
type instrumentStruct struct {
	ID         byte
	NameLength int32
	Name       string
	Pattern    [16]byte
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	header      headerStruct
	instruments map[byte]*instrumentStruct

	instrumentsOrder []byte
}

// GetHeaderPrintout will print the formatted header
// to the given buffer
func (p *Pattern) GetHeaderPrintout(buffer *bytes.Buffer) {
	buffer.WriteString(fmt.Sprintf("Saved with HW Version: %s\n", bytes.Trim(p.header.VersionBuf[:], "\x00")))
	buffer.WriteString(fmt.Sprintf("Tempo: %v\n", p.header.BPM))
}

// GetInstrumentsPrintout will print the formatted instruments
// to the given buffer
func (p *Pattern) GetInstrumentsPrintout(buffer *bytes.Buffer) {
	for _, index := range p.instrumentsOrder {
		instrument := p.instruments[index]

		var patternDiagram bytes.Buffer

		for i, val := range instrument.Pattern {
			// If multiple of 4, add |
			if i%4. == 0 {
				patternDiagram.WriteString("|")
			}

			if val == 0 {
				patternDiagram.WriteString("-")
			} else {
				patternDiagram.WriteString("x")
			}
		}

		patternDiagram.WriteString("|")

		buffer.WriteString(fmt.Sprintf("(%v) %s\t%v\n", instrument.ID, instrument.Name, patternDiagram.String()))

	}
}

func (p *Pattern) String() string {
	var printoutBuffer bytes.Buffer

	p.GetHeaderPrintout(&printoutBuffer)
	p.GetInstrumentsPrintout(&printoutBuffer)

	return printoutBuffer.String()
}
