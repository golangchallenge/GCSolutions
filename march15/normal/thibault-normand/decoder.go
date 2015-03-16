package drum

import ( 
	"fmt"
	"io/ioutil"
	"bytes"
	"encoding/binary"
	"strings"
	"errors"
)

type header struct {
	Magic 	[6]byte // 6
	Unk1	uint32 // 4
	BodyLen	uint32 // 4
	Version [32]byte // 32
	Tempo    float32 // 4
} // 50 bytes

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
// TODO: implement
func DecodeFile(path string) (*Pattern, error) {

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var h header
	buf :=  bytes.NewBuffer(data)

	// Read Magic
	err = binary.Read(buf, binary.BigEndian, &h.Magic)
	if err != nil {
		return nil, err
	}

	// Check
	if valid := checkMagic(&h.Magic); !valid {
		return nil, errors.New("Invalid file : wrong magic.")
	} 

	// Read Unknown value (padding?)
	err = binary.Read(buf, binary.BigEndian, &h.Unk1)
	if err != nil {
		return nil, err
	}

	// Splice body length
	err = binary.Read(buf, binary.BigEndian, &h.BodyLen)
	if err != nil {
		return nil, err
	}

	// Version String (32bytes null-terminated)
	err = binary.Read(buf, binary.BigEndian, &h.Version)
	if err != nil {
		return nil, err
	}
	
	// Tempo little endian float32
	binary.Read(buf, binary.LittleEndian, &h.Tempo)
	if err != nil {
		return nil, err
	}

	// Build a pattern object
	p := &Pattern{}

	// Set the version from the header
	p.Version = strings.TrimRight(string(h.Version[:]), "\x00")
	// Set the tempo from the header
	p.Tempo = h.Tempo

	// Continue reading lines
	byteRead := uint32(0)
	for byteRead <= h.BodyLen {
		// Intrument identifier
		var id uint8
		err = binary.Read(buf, binary.BigEndian, &id)
		if err != nil {
			break
		}
		byteRead++

		// Instrument name length
		var count uint32
		err = binary.Read(buf, binary.BigEndian, &count)
		if err != nil {
			break
		}
		byteRead += 4

		// Instrument name
		name := make([]byte, count)
		err = binary.Read(buf, binary.BigEndian, &name)
		if err != nil {
			break
		}
		byteRead += count

		// Steps array
		var steps [16]uint8
		err = binary.Read(buf, binary.BigEndian, &steps)
		if err != nil {
			break
		}
		byteRead += 16

		// Append the line to the pattern object
		p.Lines = append(p.Lines, drumLine{
			Number: id,
			Name: string(name),
			Hits: steps,
		})
	}

	return p, nil
}

func checkMagic(magic *[6]byte) bool {
	
	valid := [6]byte{ 0x53 /* S */, 0x50 /* P */, 0x4C /* L */, 0x49 /* I */, 0x43 /* C */, 0x45 /* E */ }

	v := true
	idx := 0
	for v {
		if idx > 5 {
			break
		}
		if valid[idx] != magic[idx] {
			v = false
		} 
		idx++
	}
 	return v
} 

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
// TODO: implement
type Pattern struct{
	Version string
	Tempo float32
	Lines []drumLine
}

func (p *Pattern) String() string {
	res := fmt.Sprintf("Saved with HW Version: %s\n", p.Version)

	t := int(p.Tempo)

	if p.Tempo - float32(t) > 0 {
		res = res + fmt.Sprintf("Tempo: %.1f", p.Tempo)
	} else {
		res = res + fmt.Sprintf("Tempo: %d", t)
	}
	for _, line := range p.Lines {
		res = res + fmt.Sprintf("\n%s", line.String())
	}
	res = res + "\n"
	return res
}

type drumLine struct {
	Number uint8
	Name string
	Hits [16]uint8  
}

func (l *drumLine) String() string {
	res := fmt.Sprintf("(%d) %s\t", l.Number, l.Name)
	for idx, hit := range l.Hits {
		if idx % 4 == 0 {
			res = res + "|"
		}
		if hit == 1 {
			res = res + "x"
		} else {
			res = res + "-"
		}
	}
	res = res + "|"
	return res
}