package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
// TODO: implement
func DecodeFile(path string) (*Pattern, error) {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	headers := make([]byte, 50)
	_, err = file.Read(headers)
	if err != nil {
		fmt.Println(err)
	}

	var instruments []*instrument
	for {
		i, err := readInstrument(file)
		if err != nil {
			break
		}
		instruments = append(instruments, i)
	}

	p := &Pattern{
		splice:      fmt.Sprintf("%s", headers[0:6]),
		version:     strings.TrimRight(fmt.Sprintf("%s", headers[14:25]), "\x00"),
		tempo:       readTempo(headers[46:50]),
		instruments: instruments,
	}

	return p, nil
}

func readInstrument(file *os.File) (*instrument, error) {
	data := make([]byte, 4)
	_, err := file.Read(data)
	if err != nil {
		return nil, err
	}
	number := readInstrumentNumber(data)

	data = make([]byte, 1)
	_, err = file.Read(data)
	if err != nil {
		return nil, err
	}
	length := readInstrumentNameLength(data)

	data = make([]byte, length)
	_, err = file.Read(data)
	if err != nil {
		return nil, err
	}
	name := readInstrumentName(data)

	data = make([]byte, 16)
	_, err = file.Read(data)
	if err != nil {
		return nil, err
	}
	beats := readInstrumentBeats(data)

	return &instrument{
		name:   name,
		number: number,
		beats:  beats,
	}, nil
}

func readInstrumentNameLength(data []byte) int8 {
	buf := bytes.NewBuffer(data)
	var val int8
	binary.Read(buf, binary.LittleEndian, &val)
	return val
}

func readInstrumentNumber(data []byte) int32 {
	buf := bytes.NewBuffer(data)
	var val int32
	binary.Read(buf, binary.LittleEndian, &val)
	return val
}

func readInstrumentName(data []byte) string {
	return fmt.Sprintf("%s", bytes.NewBuffer(data))
}

func readInstrumentBeats(data []byte) []int8 {
	buf := bytes.NewBuffer(data)
	beats := make([]int8, 16)
	for i := 0; i < 16; i++ {
		var val int8
		binary.Read(buf, binary.LittleEndian, &val)
		beats[i] = val
	}
	return beats
}

func readTempo(data []byte) float32 {
	s := bytes.NewBuffer(data)
	var val float32
	binary.Read(s, binary.LittleEndian, &val)
	return val
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
// TODO: implement
type Pattern struct {
	splice      string
	version     string
	tempo       float32
	instruments []*instrument
}

func (p *Pattern) String() string {
	version := fmt.Sprintf("Saved with HW Version: %s\n", p.version)
	tempo := fmt.Sprintf("Tempo: %g\n", p.tempo)
	instruments := ""
	for _, i := range p.instruments {
		instruments += fmt.Sprintf("%s", i)
	}
	return version + tempo + instruments
}

type instrument struct {
	number int32
	name   string
	beats  []int8
}

func (i *instrument) String() string {
	beatStr := "|"
	for index, b := range i.beats {
		if b == 0x01 {
			beatStr += "x"
		} else {
			beatStr += "-"
		}
		if ((index + 1) % 4) == 0 {
			beatStr += "|"
		}
	}

	return fmt.Sprintf("(%d) %s\t%s\n", i.number, i.name, beatStr)
}
