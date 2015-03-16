/*
	yisiper@gmail.com
	Go Challenge 1 - solution
	http://golang-challenge.com/go-challenge1/
*/

package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	defer func() {
		if e := recover(); e != nil {
			fmt.Println("Recover panic : ", e)
			return
		}
	}()
	p := &Pattern{}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	size, _ := os.Stat(path)
	b := make([]byte, size.Size())

	// full read
	err = binary.Read(file, binary.LittleEndian, &b)
	if err != nil {
		return nil, err
	}

	// splice - version - tempo - track
	if bytes.Compare([]byte("SPLICE"), b[0:6]) != 0 {
		return nil, errors.New("SPLICE Not Found")
	}

	//fmt.Println(hex.Dump(b))
	var tid uint8
	var tname string

	var idpos uint32     // starting position for first id
	var length uint32    // name length
	var lengthpos uint32 // starting position for name length
	var stepPos uint32   // starting position for step
	var stepSize uint32  // track step size = 16
	var nextidPos uint32 // starting position for next id
	var bsize uint32     // binary size
	var isize uint32     // byte size

	stepSize = 16 // 16 bytes step size
	idpos = 50
	lengthpos = 51
	isize = uint32(len(b))

	for true {
		// read the id
		buf := bytes.NewReader(b[idpos : idpos+1])
		binary.Read(buf, binary.LittleEndian, &tid)

		// handle binary size ?
		buf = bytes.NewReader(b[10:14])
		binary.Read(buf, binary.BigEndian, &bsize)

		// Why ?
		buf = bytes.NewReader(b[lengthpos : lengthpos+4])
		binary.Read(buf, binary.BigEndian, &length)

		if length > bsize || length > isize {
			//fmt.Println(length, lengthpos, stepPos)
			break
		}
		stepPos = lengthpos + 4 + length
		nextidPos = stepPos + stepSize

		// name
		tname = string(b[lengthpos+4 : stepPos])

		p.version = string(bytes.Trim(b[14:45], string(0x00)))
		p.tempo = float32frombytes(b[46:50])
		p.track = append(p.track, track{id: tid, name: tname, step: b[stepPos:nextidPos]})

		idpos = nextidPos
		lengthpos = idpos + 1
		if idpos >= isize || idpos > bsize {
			break
		}
	}
	return p, nil
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	version string
	tempo   float32
	track   []track
}

type track struct {
	id   uint8
	name string
	step []byte
}

func (p *Pattern) String() string {
	if p == nil {
		return ""
	}

	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintln("Saved with HW Version:", p.version))
	buffer.WriteString(fmt.Sprintln("Tempo:", p.tempo))

	for _, j := range p.track {
		buffer.WriteString(fmt.Sprint("(", strconv.Itoa(int(j.id)), ") ", j.name, "\t"))
		step := 0
		for i := 0; i < 16; i++ {
			if math.Mod(float64(step), 4) == 0 {
				//buffer.WriteByte(0x7C)
				buffer.WriteString("|")
			}
			if bytes.Compare(j.step[i:i+1], []byte{0x00}) == 0 {
				buffer.WriteString("-")
			} else if bytes.Compare(j.step[i:i+1], []byte{0x01}) == 0 {
				buffer.WriteString("x")
			} else {
				// incase if the byte is not 0x00 0x01
				buffer.WriteString(" ")
			}
			step = step + 1
			if i == 15 {
				buffer.WriteString("|")
				//buffer.WriteByte(0x7C)
			}
		}
		buffer.WriteString("\n")
	}
	return buffer.String()
}

func float32frombytes(bytes []byte) float32 {
	bits := binary.LittleEndian.Uint32(bytes)
	float := math.Float32frombits(bits)
	return float
}

/*
 h 00 01 02 03 04 05 06 07 08 09 0A 0B 0C 0D 0E 0F
00 53 50 4C 49 43 45 00 00 00 00 00 00 00 C5 30 2E
10 38 30 38 2D 61 6C 70 68 61 00 00 00 00 00 00 00
20 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00
30 F0 42 00 00 00 00 04 6B 69 63 6B 01 00 00 00 01
40 00 00 00 01 00 00 00 01 00 00 00 01 00 00 00 05
50 73 6E 61 72 65 00 00 00 00 01 00 00 00 00 00 00
60 00 01 00 00 00 02 00 00 00 04 63 6C 61 70 00 00
70 00 00 01 00 01 00 00 00 00 00 00 00 00 00 03 00
80 00 00 07 68 68 2D 6F 70 65 6E 00 00 01 00 00 00
90 01 00 01 00 01 00 00 00 01 00 04 00 00 00 08 68
A0 68 2D 63 6C 6F 73 65 01 00 00 00 01 00 00 00 00
B0 00 00 00 01 00 00 01 05 00 00 00 07 63 6F 77 62
C0 65 6C 6C 00 00 00 00 00 00 00 00 00 00 01 00 00
D0 00 00 00

desc offset in hex :
0 - 9	: text => SPLICE
A - D	: binary Size
E - 2D	: version
2E - 31 : tempo
32		: id
33 - 36	: name length
37 - 3A	: name
3B - 4A : step

continue the same pattern for id, name length, name, step
*/
