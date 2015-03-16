package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
)

const offsetdatastart = 14

type Pattern struct {
	header patternheader
	data   patterndata
}

type patternheader struct {
	headerName [6]byte
	lengthFile int64
}

type patterndata struct {
	version [32]byte
	tempo   float32
	track   []patterntrack
}

type patterntrack struct {
	id         int32
	lengthName int8
	name       []byte
	steps      [16]byte
}

//Print Pattern
func (this Pattern) String() string {

	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("Saved with HW Version: %v\nTempo: %v\n", string(bytes.Trim(this.data.version[:], "\x00")), this.data.tempo))
	for sp := range this.data.track {
		buffer.WriteString(fmt.Sprintf("(%v) %v\t", this.data.track[sp].id, string(this.data.track[sp].name[:])))
		for i := 0; i < 16; i++ {
			if i%4 == 0 {
				buffer.WriteString("|")
			}

			if this.data.track[sp].steps[i] == 0 {
				buffer.WriteString("-")
			} else {
				buffer.WriteString("x")
			}
		}
		buffer.WriteString("|\n")

	}
	return buffer.String()

}

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	var p Pattern
	buffer, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = decodeHeader(buffer, &p)
	if err != nil {
		return nil, err
	}
	err = decodeData(buffer, &p)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

//Decode Data
func decodeData(buf []byte, pattern *Pattern) error {

	var track patterntrack

	if len(buf[offsetdatastart:]) < int(pattern.header.lengthFile) {
		return errors.New("Data of splicefile is too short")
	}

	buffer := bytes.NewBuffer(buf[offsetdatastart : int(pattern.header.lengthFile)+offsetdatastart])
	if err := binary.Read(buffer, binary.LittleEndian, &pattern.data.version); err != nil {
		return err
	}
	if err := binary.Read(buffer, binary.LittleEndian, &pattern.data.tempo); err != nil {
		return err
	}

	pattern.data.track = make([]patterntrack, 0)

	for buffer.Len() > 0 {
		if err := binary.Read(buffer, binary.LittleEndian, &track.id); err != nil {
			return err
		}
		if err := binary.Read(buffer, binary.LittleEndian, &track.lengthName); err != nil {
			return err
		}
		track.name = make([]byte, track.lengthName)
		if err := binary.Read(buffer, binary.LittleEndian, &track.name); err != nil {
			return err
		}
		if err := binary.Read(buffer, binary.LittleEndian, &track.steps); err != nil {
			return err
		}
		pattern.data.track = append(pattern.data.track, track)
	}

	return nil
}

//Decode Header
func decodeHeader(buf []byte, pattern *Pattern) error {

	if len(buf) < offsetdatastart {
		return errors.New("Header of splicefile is too short")
	}

	buffer := bytes.NewBuffer(buf)
	if err := binary.Read(buffer, binary.LittleEndian, &pattern.header.headerName); err != nil {
		return err
	}
	if err := binary.Read(buffer, binary.BigEndian, &pattern.header.lengthFile); err != nil {
		return err
	}

	return nil
}
