package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	var err error
	var count int

	// fileInfo will take the first 14 bytes
	// This includes the "SPLICE" header
	// and the number of bytes to read
	fileInfo := make([]byte, 14)

	// Opening the file and error handling
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	// Reading the fileInfo and error handling
	count, err = file.Read(fileInfo)
	if err != nil {
		return nil, err
	}
	if count < 14 {
		return nil, errors.New("Binary file too short.")
	}

	// data will be the holder for the remaining bytes
	// on the splice file. It stores the number of bytes
	// indicated on byte 14, which told us how many
	// bytes we should be reading
	data := make([]byte, fileInfo[13])

	// Setting the start point, reading the file, and error handling
	file.Seek(14, 0)
	count, err = file.Read(data)
	if err != nil {
		return nil, err
	}
	file.Close()

	// A drum pattern must have at least 36 bytes
	// which only includes the Version and Tempo
	if count < 36 {
		return nil, errors.New("Binary file too short.")
	}

	// Decode the version from the first 32 bytes
	version := DecodeVersion(data[0:32])

	// Decode the tempo from the next 4 bytes
	tempo, err := DecodeTempo(data[32:36])
	if err != nil {
		return nil, err
	}

	// Decode the instruments from the remaining bytes
	instruments := DecodeInstruments(data[36:count])

	// Buffer to hold the building output
	var buf bytes.Buffer

	// Formatting the output
	buf.WriteString(version)
	buf.WriteString(tempo)
	buf.WriteString(instruments)

	// Initializing the Pattern struct with the parsed pattern
	p := &Pattern{buf.String()}
	return p, nil
}

// DecodeInstrument decodes the instrument lines from the corresponding
// byte slices and returns a string of the instrument output
func DecodeInstruments(data []byte) string {
	// Buffer to hold the building output
	var buf bytes.Buffer
	count := 0
	for count < len(data) {
		// Writes the Instrument Code to the buffer
		buf.WriteString(fmt.Sprintf("(%d) ", data[count]))

		// Assigns chars to the character count byte
		// which tells us how many characters are in
		// an instrument's name
		chars := int(data[count+4])

		// Sends the remaining bytes for this instrument
		// to the DecodeOneInstrument for specific decoding
		// from the start of the name (character count included)
		// through the end of the beat
		instrumentLine := DecodeOneInstrument(data[count+4 : count+chars+21])
		buf.WriteString(instrumentLine)

		// Setting up for the next instrument in this loop
		count += chars + 21
	}
	// Returning a stringified version of the output
	return buf.String()
}

// DecodeOneInstrument decodes a single instrument from the corresponding
// byte slices and returns a string of the instrument output to the
// DecodeInstruments function
func DecodeOneInstrument(data []byte) string {
	// Buffer to hold the building output
	var buf bytes.Buffer

	// Decoding the instrument name
	nameChars := int(data[0]) + 1
	buf.WriteString(string(data[1:nameChars]))
	buf.WriteString("\t")

	// Decoding the beat
	for i := 0; i < 16; i++ {
		if i%4 == 0 {
			buf.WriteString("|")
		}
		if data[nameChars+i] == 1 {
			buf.WriteString("x")
		} else {
			buf.WriteString("-")
		}
	}
	buf.WriteString("|\n")

	// Returning a stringified version of the output
	return buf.String()
}

// DecodeTempo decodes the tempo from the corresponding
// byte slices and returns a string of the tempo output
func DecodeTempo(data []byte) (string, error) {
	// Decoding the tempo and storing it in tempo
	var tempo float32
	err := binary.Read(bytes.NewReader(data), binary.LittleEndian, &tempo)
	if err != nil {
		return "", err
	}

	// Buffer to hold the building output
	var buf bytes.Buffer

	// Formatting the Tempo line
	buf.WriteString("Tempo: ")
	buf.WriteString(fmt.Sprintf("%g", tempo))
	buf.WriteString("\n")

	// Returning a stringified version of the output
	return buf.String(), nil
}

// DecodeVersion decodes the version information from the
// corresponding byte slices and returns a string of the
// version output
func DecodeVersion(data []byte) string {
	// Finds the first empty byte and sets
	// end equal to that byte's position
	end := 0
	for i := 0; int(data[i]) != 0; i++ {
		end = i + 1
	}

	// Buffer to hold the building output
	var buf bytes.Buffer

	// Formatting the Version line
	buf.WriteString("Saved with HW Version: ")
	buf.WriteString(string(data[0:end]))
	buf.WriteString("\n")

	// Returning a stringified version of the output
	return buf.String()
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	output string
}

// String function overwrites the default for the Pattern struct
// and returns the parsed pattern as a string
func (p Pattern) String() string {
	return p.output
}
