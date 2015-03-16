package drum

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"strings"
)

const (
	SpliceSize  = 6  //Number of bytes of "SPLICE"
	HeadSize    = 13 //Number of bytes of the header
	VersionSize = 32 //Number of bytes of the version
	TempoSize   = 4  //Number of bytes of the tempo
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	//Open the file and defers its closing.
	file, err_open := os.Open(path)
	if err_open != nil {
		return nil, err_open
	}
	defer file.Close()

	//Put the data of the file in a []byte
	data, err_extract := ioutil.ReadAll(file)

	if err_extract != nil {
		return nil, err_extract
	}

	//extract the headers and return an error if the file is not a drum machine.
	if !extractHeader(&data) {
		return nil, errors.New("Wrong file format")
	}

	size := extractSize(&data)

	if size < 0 {
		return nil, errors.New("Error in the file.")
	}

	version := extractVersion(&data)

	tempo := extractTempo(&data)

	if tempo < 0 {
		return nil, errors.New("Negative tempo.")
	}

	lines := make(Lines, size/25)

	for i := 0; i < len(lines); i++ {
		lines[i] = *(extractLine(&data))
	}

	return &Pattern{version, tempo, lines}, nil
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	version string
	tempo   float32
	lines   Lines
}

// Line is the representation of a the 16-times
// measure of an instrument
type Line struct {
	id      int
	sound   string
	measure [16]bool
}

// Lines is an array of Line
type Lines []Line

// extractHeader returns true if the header is SPLICE
// and else, returns false. It also removes the 13 header
// byte from data.
func extractHeader(data *[]byte) bool {
	res := string((*data)[:SpliceSize])
	*data = (*data)[HeadSize:]
	return res == "SPLICE"
}

// extractSize returns the size in number of bytes
// of the lines. It also removes the byte indicating the size
// of the lines, the version and the tempo from data.
func extractSize(data *[]byte) (res int) {
	res = int((*data)[0]) - 36
	*data = (*data)[1:]
	return
}

// extractVersion returns a string containing the version
// used to create the binary file. It also removes the 32
// bytes indicating the version from data.
func extractVersion(data *[]byte) (res string) {
	res = strings.TrimRight(string((*data)[:VersionSize]), string(0x00))
	*data = (*data)[VersionSize:]
	return
}

// extractTempo returns a float32 containing the tempo of the
// music. It also removes the 4 bytes indicating the tempo
// from data.
func extractTempo(data *[]byte) (res float32) {
	bits := binary.LittleEndian.Uint32((*data)[:TempoSize])
	res = math.Float32frombits(bits)
	*data = (*data)[TempoSize:]
	return
}

// extractLine returns a the next Line read from data and
// its binary representation from data.
func extractLine(data *[]byte) *Line {
	// Extract the id of the Line
	id := int(binary.BigEndian.Uint16([]byte{0x00, (*data)[0]}))

	*data = (*data)[1:]

	// Extract the number of characters in the name of the instrument
	nb_char := int(binary.BigEndian.Uint32((*data)[:4]))

	*data = (*data)[4:]

	// Extract the name of the instrument
	name := string((*data)[:nb_char])

	*data = (*data)[nb_char:]

	// Extract the measure.
	var measure [16]bool
	for i := 0; i < 16; i++ {
		measure[i] = ((*data)[i] == 0x01)
	}

	*data = (*data)[16:]

	return &Line{id, name, measure}
}

// String returns the string representation of Lines.
func (lines Lines) String() string {
	res := ""

	for i := 0; i < len(lines); i++ {
		// Adds the id and the name of the sound to the result.
		res += fmt.Sprintf("(%v) %v\t", lines[i].id, lines[i].sound)

		// Adds the the x and - representation of the measure.
		for j := 0; j < 16; j += 4 {
			res += "|"
			for k := 0; k < 4; k++ {
				if lines[i].measure[j+k] {
					res += "x"
				} else {
					res += "-"
				}
			}
		}
		res += fmt.Sprintf("|\n")
	}

	return res
}

// String returns the string representation of a pattern.
func (p *Pattern) String() string {
	return fmt.Sprintf("Saved with HW Version: %v\nTempo: %v\n%v",
		p.version,
		p.tempo,
		p.lines.String())
}
