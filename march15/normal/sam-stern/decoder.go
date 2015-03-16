package drum

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"regexp"
)

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	hwVersion       string
	tempo           float32
	instrumentNames []string
	instrumentIds   map[string]int
	instrumentBeats map[string][]byte
}

// Regex for 'word' characters, which are letters, numbers, dashes, spaces
// and periods.
var wordPattern = regexp.MustCompile("[a-zA-Z0-9\\.\\s\\-]")

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {

	// Initialize members of the Pattern
	p := &Pattern{
		instrumentIds:   make(map[string]int),
		instrumentBeats: make(map[string][]byte),
	}

	// Read the file into a []byte
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println(err)
		return p, err
	}

	// Explanation of file format
	// 6B   SPLICE
	// 8B   padding
	// 32B  Hardware Version + padding
	// 4B   Tempo
	// *    instrument names, ids, and beats...
	headerData := dat[0:6]
	hwVersionData := dat[(6 + 8):(6 + 8 + 32)]
	tempoData := dat[(6 + 8 + 32):(6 + 8 + 32 + 4)]
	instrumentData := dat[(6 + 8 + 32 + 4):]

	// Return an error if the first 6 characters are not "SPLICE"
	if string(headerData) != "SPLICE" {
		return p, errors.New("first 6 characters must be 'SPLICE'")
	}

	// Extract the hardware version as a string
	hwVersionStr, _ := ExtractFirstWord(hwVersionData)
	p.hwVersion = hwVersionStr

	// Extract the tempo as a float
	tempoFloat := BytesToFloat(tempoData)
	p.tempo = tempoFloat

	// beatNumber is the # of beats we have seen for the current instrument
	beatNumber := 0

	// inBeatMode is true if looking for beats (0s and 1s), false if looking for instrument name
	inBeatMode := false

	// Preprocessing: some files may accidentally contain multiple beat records,
	// we only care about the first.  So if we see SPLICE in the instrument data,
	// cut it off there
	endDataIndex := len(instrumentData)
	for i := 0; i < len(instrumentData)-6; i++ {
		// Check for SPLICE, first checking for 'S' to save the string allocation
		if instrumentData[i] == 'S' && string(instrumentData[i:i+6]) == "SPLICE" {
			endDataIndex = i
			break
		}
	}
	instrumentData = instrumentData[0:endDataIndex]

	var currentInstrument string
	for i := 0; i < len(instrumentData); {
		if !inBeatMode {
			// Extract next word from the data
			str, j := ExtractFirstWord(instrumentData[i:])
			currentInstrument = str
			i += j

			// Check bounds after increment
			if i >= len(instrumentData) {
				break
			}

			// Found an instrument name, append it to the list and initialize
			// its array of beats to size 16
			p.instrumentNames = append(p.instrumentNames, currentInstrument)
			p.instrumentBeats[currentInstrument] = make([]byte, 16)

			// Character 5 before the name was the ID
			idInd := i - len(currentInstrument) - 5
			p.instrumentIds[currentInstrument] = int(instrumentData[idInd])

			// Reset the beat to 0 with each new word found
			beatNumber = 0
			inBeatMode = true
		} else {
			// In beat mode
			if beatNumber < 16 {
				// 0 and 1 represent beats
				if instrumentData[i] <= 1 {
					// Store beat in map, advance to next beat
					p.instrumentBeats[currentInstrument][beatNumber] = instrumentData[i]
					beatNumber++
				}

				// Move to next byte
				i++
			} else {
				// Found 16 beats, leave beat mode
				inBeatMode = false

				// There is a 5-byte gap between the end of 16 beats
				// and the start of the next instrument name
				i += 5
			}
		}
	}

	return p, nil
}

// ExtractFirstWord takes an array of bytes and finds the first
// continuous stretch of word characters where a word character
// is a letter, number, space, dash or period. Returns the string
// found and the index of the first letter after the string. If
// no word is found, the blank string "" is returnd.
func ExtractFirstWord(data []byte) (string, int) {
	// Have we found a word character yet?
	foundWordChar := false

	// What is the index of the first word character?
	var firstWordChar int

	for i := 0; i < len(data); i++ {
		isWordChar := wordPattern.MatchString(string(data[i]))
		if !foundWordChar && isWordChar {
			foundWordChar = true
			firstWordChar = i
		}

		// Return after we find a non-word character (after a series of word
		// characters)
		if foundWordChar && !isWordChar {
			return string(data[firstWordChar:i]), i
		}
	}

	if foundWordChar {
		// All characters were word characters, the whole thing is a word
		return string(data), len(data)
	}

	// No word characters at all, return blank string
	return "", len(data)
}

// BytesToFloat takes an array of 4 bytes and converts it into a
// 32-bit float. The bytes are considered in the opposite order that
// they are stored in the array (arr[3] first, arr[0] last) and converted
// using the IEEE 754 floating point number specification.
func BytesToFloat(data []byte) float32 {
	// Convert to 4 ints
	i0, i1, i2, i3 := int(data[0]), int(data[1]), int(data[2]), int(data[3])

	// Reverse, shift, and OR to get one binary number
	u32 := uint32(i3<<24 | i2<<16 | i1<<8 | i0)

	// Convert to Float32
	return math.Float32frombits(u32)
}

// Print a pattern in the required human-readable format
func (p Pattern) String() string {
	var buffer bytes.Buffer

	// Print the hardward version
	buffer.WriteString(fmt.Sprintf("Saved with HW Version: %s\n", p.hwVersion))

	// The tempo should be presented with no trailing zeros
	buffer.WriteString(fmt.Sprintf("Tempo: %.4g\n", p.tempo))

	// Print all the instrument ids, names, and beat patterns
	for i := 0; i < len(p.instrumentNames); i++ {
		instrument := p.instrumentNames[i]

		// Print instrument name and ID
		buffer.WriteString(fmt.Sprintf("(%d) %s\t", p.instrumentIds[instrument], instrument))

		// Print all beats for the instrument, adding a '|' every 4th beat and before
		// the first and last beats
		for i, elm := range p.instrumentBeats[instrument] {
			if i%4 == 0 {
				buffer.WriteString("|")
			}

			if elm == 0 {
				buffer.WriteString("-")
			} else if elm == 1 {
				buffer.WriteString("x")
			}

		}

		buffer.WriteString("|\n")
	}

	return buffer.String()
}
