package drum

import (
	"fmt"
	"io/ioutil"
	"strings"
	"strconv"
	"errors"
	"encoding/binary"
	"bytes"
)

func DecodeFile(path string) (*Pattern, error) {
	p := &Pattern{}
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	p, err = PatternData(file)
	return p, err
}

func PatternData(data []byte) (*Pattern, error) {
	po := &Pattern{}
	po.version = strings.TrimRight(string(data[14:45]), "\x00")

	real_l := len(data)
	l := int(data[13])+14
	if real_l<l {
		return nil, errors.New("decoder: Wrong format")
	}

	buf := bytes.NewReader(data[46:50])
	err := binary.Read(buf, binary.LittleEndian, &po.tempo)
	if err != nil {
		fmt.Println("binary.Write failed:", err)
	}

	for i:=50;i<l;i++ {
		if i+20>l {
			return nil, errors.New("decoder: Wrong format")
		}
		t := Track{}
		t.id = int(data[i])
		i += 4
		name_length := int(data[i])
		if i+name_length+16>l {
			return po, nil
		}
		i += 1
		t.name = string(data[i:(i+name_length)])
		i += name_length
		for j:=0;j<16;j++ {
			t.steps[j] = int(data[i+j])==1
		}
		i += 15
		po.tracks = append(po.tracks, t)
	}
	return po, nil
}

func (p Pattern) String() string {
	echo := fmt.Sprintf("Saved with HW Version: %v", p.version)
	echo += "\n"
	echo += fmt.Sprintf("Tempo: %v", p.tempo)
	echo += "\n"

	for _, t := range p.tracks {
		echo += "(" + strconv.Itoa(t.id) + ") " + t.name + "	"
		for s:=0;s<16;s++ {
			if s%4==0 {
				echo += "|"
			}
			if t.steps[s] {
				echo += "x"
			} else {
				echo += "-"
			}
		}
		echo += "|"
		echo += "\n"
	}
	return echo
}

type Track struct {
	id int
	name string
	steps [16]bool
}

type Pattern struct{
	version string
	tempo float32
	tracks []Track
}
