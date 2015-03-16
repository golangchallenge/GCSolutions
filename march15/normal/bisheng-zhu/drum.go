//auther: doun@github.com
// Package drum  implement the decoding of .splice drum machine files.
// from golang-challenge.com/go-challenge1/ for more information

package drum

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"strings"
)

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
const (
	len_end_pos   = 4
	ver_end_pos   = 0x2e - header_end_pos
	tempo_end_pos = ver_end_pos + 4
)

var (
	FileFormatERR  = errors.New("File format err")
	WrongHeaderERR = errors.New("Wrong header")
)

type track struct {
	id    uint8
	name  string
	steps [16]byte
}
type Pattern struct {
	Version string
	Tempo   float32
	Tracks  []track
}

/*
data [] without header
*/
func (s *Pattern) UnmarshalBinary(data []byte) error {
	//version
	v_buf := data[len_end_pos:ver_end_pos]
	s.Version = strings.TrimFunc(string(v_buf), func(r rune) bool {
		if r == 0x0 {
			return true
		}
		return false
	})
	//tempo
	s.Tempo = math.Float32frombits(
		binary.LittleEndian.Uint32(data[ver_end_pos:tempo_end_pos]))
	var cur_pos = tempo_end_pos
	var nm_len int
	for {
		if cur_pos >= len(data) {
			break
		}
		var st track
		//id
		st.id = uint8(data[cur_pos])
		cur_pos++
		//name
		// the length
		nm_len = int(binary.BigEndian.Uint32(data[cur_pos : cur_pos+4]))
		cur_pos += 4
		// the real name bytes
		st.name = string(data[cur_pos : cur_pos+int(nm_len)])
		cur_pos += nm_len
		//kicks
		n := copy(st.steps[:], data[cur_pos:cur_pos+16])
		if n != 16 {
			return FileFormatERR
		}
		cur_pos += 16
		s.Tracks = append(s.Tracks, st)
	}
	return nil

}

func (s Pattern) String() string {
	var str string
	str += fmt.Sprintf("Saved with HW Version: %s\nTempo: %v\n",
		s.Version, s.Tempo)
	for _, st := range s.Tracks {
		str += fmt.Sprintf("%v", st)
	}
	return str
}

func (s track) String() string {
	return fmt.Sprintf("(%d) %s	%v", s.id, s.name, steps(s.steps))
}

type step byte

func (s step) String() string {
	if byte(s) == 0 {
		return "-"
	} else {
		return "x"
	}
}

type steps [16]byte

func (s steps) String() string {
	var str string
	kicks := [16]byte(s)
	for i, single := range kicks {
		if i%4 == 0 {
			str += "|"
		}
		str += fmt.Sprintf("%v", step(single))
	}
	str += "|\n"
	return str
}
