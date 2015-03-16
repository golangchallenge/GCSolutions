package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
)

// Some custom types so we can print them as strings
type toolName [13]byte
type toolVersion [32]byte
type steps [16]byte

// Pattern represents a decoded splice file
type Pattern struct {
	Header      header
	Instruments []instrument
}

// header appears at the start of every splice file
type header struct {
	Tool     toolName
	DataSize uint8
	Version  toolVersion
	Tempo    float32
}

type instrument struct {
	Number uint32
	Name   []byte
	Steps  steps
}

func (n toolName) String() string {
	return cleanBytesToString(n[0:])
}

func (v toolVersion) String() string {
	return cleanBytesToString(v[0:])
}

// OnBeat returns true if the step is playing for that beat
// Used by the wav writer
func (s steps) onBeat(beat int) bool {
	return s[beat] != 0
}

func (s steps) String() string {
	out := "|"
	for i, e := range s {
		if e != 0 {
			out += "x"
		} else {
			out += "-"
		}
		if i%4 == 3 {
			out += "|"
		}
	}
	return out
}

/*
// Earlier attempt at using a byte array, but string appends are fast enough, so
// I prefer the above
func (s steps) String() string {
	out := [21]byte{}
	out[0], out[5], out[10], out[15], out[20] = '|', '|', '|', '|', '|'
	c := 0
	for i, e := range out {
		if e != '|' {
			if s[i-c] != 0 {
				out[i] = 'x'
			} else {
				out[i] = '-'
			}
		} else {
			c++
		}
	}
	return string(out[0:])
}
*/

func (i instrument) String() string {
	return fmt.Sprintf("(%d) %s\t%s", i.Number, i.Name, i.Steps)
}

// cleanBytesToString converts our byte splice to a string, splits it on any nulls and returns the first part of the string
func cleanBytesToString(s []byte) string {
	return strings.SplitN(string(s), "\000", 2)[0]
}

// String prints out our pattern as our test expects
func (p Pattern) String() string {
	out := fmt.Sprintf("Saved with HW Version: %s\n", p.Header.Version)
	out += fmt.Sprintf("Tempo: %g\n", p.Header.Tempo)
	for _, i := range p.Instruments {
		out += i.String() + "\n"
	}
	return out
}

func decode(r io.Reader) (*Pattern, error) {
	p := &Pattern{}
	err := binary.Read(r, binary.LittleEndian, &p.Header)
	if err != nil {
		return nil, err
	}

	// Subtract size of our tool version and tempo values
	bytesRemaining := int(p.Header.DataSize) - 36

	var data []byte
	// Reuse our existing byte array if we were passed one
	d, ok := r.(*bytes.Buffer)
	if ok {
		data = d.Bytes()
	} else {
		data = make([]byte, bytesRemaining)
	}
	bytesRead, err := r.Read(data)

	if bytesRead > bytesRemaining {
		//log.Printf("Read more bytes then expected, potentially corrupt splice file")
		data = data[0:bytesRemaining]
	}
	if bytesRead < bytesRemaining {
		return nil, fmt.Errorf("Unable to read whole splice file: expected another %d bytes", bytesRemaining-bytesRead)
	}
	if err != nil {
		return nil, err
	}
	offset := 0

	for offset < len(data) {
		// Instrument number
		number := binary.LittleEndian.Uint32(data[offset : offset+4])
		offset += 4

		// Length of our instrument name
		nameLen := int(data[offset : offset+1][0])
		offset++
		// Variable length strings are often written poorly, so
		// validate that our buffer has enough data for the string,
		// and the following 16 bytes(the steps in the measure)
		if offset+nameLen > bytesRead-16 {
			return p, fmt.Errorf("Instrument name extends beyond buffer, %d %d %d %d", offset+nameLen, len(data), bytesRead-16, nameLen)
		}

		// A slice pointing to our instument name
		nameBuf := data[offset : offset+nameLen]
		offset += nameLen

		i := instrument{Number: number, Name: nameBuf}
		copy(i.Steps[:], data[offset:offset+16])
		offset += 16

		p.Instruments = append(p.Instruments, i)
	}
	return p, nil
}

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Error opening fixture: %+v", err)
	}

	return decode(bytes.NewBuffer(data))
}
