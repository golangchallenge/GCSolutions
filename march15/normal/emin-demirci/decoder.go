package drum

import (
	"bytes"
	"container/list"
	"encoding/binary"
	"math"
	"os"
	"strconv"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
// TODO: implement
func DecodeFile(path string) (*Pattern, error) {
	p := &Pattern{}
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		panic(err)
	}

	var size = info.Size()
	var data = make([]byte, size)
	err = binary.Read(f, binary.BigEndian, data)
	if err != nil {
		panic(err)
	}

	FromBytes(p, data)

	return p, nil
}

func (p Pattern) String() string {

	var buffer bytes.Buffer
	buffer.WriteString("Saved with HW Version: ")
	buffer.WriteString(p.Version)
	buffer.WriteString("\n")
	buffer.WriteString("Tempo: ")
	buffer.WriteString(strconv.FormatFloat(float64(p.Tempo), 'f', -1, 32))
	buffer.WriteString("\n")

	for e := p.Samples.Front(); e != nil; e = e.Next() {
		buffer.WriteString("(")
		buffer.WriteString(strconv.FormatInt(int64(e.Value.(*Sample).Id), 10))
		buffer.WriteString(")")
		buffer.WriteString(" ")
		buffer.WriteString(e.Value.(*Sample).Name)
		buffer.WriteString("\t")
		// fmt.Println(len(e.Value.(*Sample).Kicks))

		for i := 0; i < 16; i++ {
			if i%4 == 0 {
				buffer.WriteString("|")
			}
			if e.Value.(*Sample).Kicks[i] == 1 {
				buffer.WriteString("x")
			} else {
				buffer.WriteString("-")
			}
		}
		buffer.WriteString("|\n")
	}

	return buffer.String()
}

func FromBytes(p *Pattern, data []byte) *Pattern {

	var buffer = bytes.NewBuffer(data)
	p.Splice = string(buffer.Next(6))
	p.DataSize = ReadUint64(buffer.Next(8))
	p.Version = string(bytes.Trim(buffer.Next(32), "\x00"))
	p.Tempo = Float32frombytes(buffer.Next(4))
	p.Samples = list.New()
	for buffer.Len() > 0 {
		if string(buffer.Next(1)) == "S" {
			return p
		} else {
			buffer.UnreadByte()
		}
		var id = buffer.Next(1)[0]
		var nameSize = int(ReadUint32(buffer.Next(4)))
		var name = string(buffer.Next(nameSize))
		var kicks = buffer.Next(16)
		p.Samples.PushBack(&Sample{id, name, kicks})
	}
	return p
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
// TODO: implement
type Pattern struct {
	Splice   string
	DataSize uint64
	Version  string
	Tempo    float32
	Samples  *list.List
}

type Sample struct {
	Id    byte
	Name  string
	Kicks []byte
}

func ReadUint64(buffer []byte) uint64 {
	return binary.BigEndian.Uint64(buffer)
}

func ReadUint32(buffer []byte) uint32 {
	return binary.BigEndian.Uint32(buffer)
}

func Float32frombytes(bytes []byte) float32 {
	bits := binary.LittleEndian.Uint32(bytes)
	float := math.Float32frombits(bits)
	return float
}
