package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	p := &Pattern{}

	// Get usable `version` string.
	// Must be better way to do this...
	usable_version_length := 0
	for i := 14; i < 25; i++ {
		if contents[i] == 0 {
			break
		}

		usable_version_length += 1
	}
	p.version = string(contents[14 : 14+usable_version_length])

	// Get the `tempo`. `tempo` is recorded as a float32.
	buf := bytes.NewReader(contents[46:50])
	if err := binary.Read(buf, binary.LittleEndian, &p.tempo); err != nil {
		panic(err)
	}

	// Attach the rest of the `score` to the Pattern.
	p.score = contents[50:len(contents)]

	return p, nil
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	version string
	tempo   float32
	score   []byte
}

func (p Pattern) String() string {
	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf("Saved with HW Version: %s\n", p.version))
	buffer.WriteString(fmt.Sprintf("Tempo: %g\n", p.tempo))

	cur := 0
	instrument_length := 0

	// For storing the `id` of each instrument.
	var id uint32

	for {
		// Read the `id` into a uint32
		buf := bytes.NewReader(p.score[cur : cur+4])
		if err := binary.Read(buf, binary.LittleEndian, &id); err != nil {
			panic(err)
		}
		cur += 4

		instrument_length = int(p.score[cur])
		cur += 1

		// Because possible fuck-ups in the encoding for the length of
		// the instrument.
		if cur+instrument_length > len(p.score) {
			break
		}

		// Start adding the `score` to the buffer.

		// Add the `id` and `instrument` name to buffer.
		buffer.WriteString(fmt.Sprintf("(%d) %s\t", id, string(p.score[cur:cur+instrument_length])))
		cur += instrument_length

		// Adds each instiruments score to buffer.
		for i := 0; i < 16; i++ {
			if i%4 == 0 {
				buffer.WriteString("|")
			}
			if p.score[cur+i] == 1 {
				buffer.WriteString("x")
			} else {
				buffer.WriteString("-")
			}
		}
		buffer.WriteString("|\n")
		cur += 16

		if cur >= len(p.score) {
			break
		}
	}

	return buffer.String()
}
