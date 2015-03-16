// Package drum is supposed to implement the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// Info structure of the individual drum information.
type Info struct {
	id    uint32
	name  []uint8
	steps [16]uint8
}

// DecodeBeat decodes the individual drum information.
func DecodeBeat(r io.Reader, p *Pattern) {
	var err error

	for {
		var d Info
		var nameSize uint8

		if err = binary.Read(r, binary.LittleEndian, &d.id); err != nil {
			if err != io.EOF {
				fmt.Println("Failed to read drum id")
			}
			break
		}

		if err = binary.Read(r, binary.LittleEndian, &nameSize); err != nil {
			fmt.Println("Failed to read drum nameSize")
			break
		}

		d.name = make([]uint8, nameSize)
		if err = binary.Read(r, binary.LittleEndian, &d.name); err != nil {
			fmt.Println("Failed to read drum name")
			break
		}

		if err = binary.Read(r, binary.LittleEndian, &d.steps); err != nil {
			fmt.Println("Failed to read drum steps information")
			break
		}

		p.tracks = append(p.tracks, d)
	}

	if err != io.EOF {
		fmt.Println(err)
	}
}

// String creates a formatted string of individual drum information.
func (data Info) String() string {
	var buff bytes.Buffer

	// start string with drum number and name
	buff.WriteString(fmt.Sprint("(", data.id, ") ", string(data.name), "\t"))

	// loop through indivdual drum information and append to string
	for i := 0; i < len(data.steps); i++ {
		if i%4 == 0 {
			buff.WriteString("|")
		}

		if data.steps[i] == 1 {
			buff.WriteString("x")
		} else {
			buff.WriteString("-")
		}
	}
	buff.WriteString("|\n")

	return buff.String()
}
