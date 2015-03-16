package drum

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"
)

type track struct {
	id    int
	name  string
	steps string
}

type Header struct {
	Signature [12]byte
	Reserved1 [2]byte
	Hwversion [32]byte
	Tempo     float32
}

func boolString(notes []byte) string {
	var t string = ""

	for j := 0; j <= 15; j++ {
		if j%4 == 0 {
			t += "|"
		}

		if int(notes[j]) == 0 {
			t += "-"
		} else {
			t += "x"
		}
	}
	return t + "|"
}

func decodeDrum(filename string) (*Header, *[]track) {
	var tracks []track
	var id uint32
	var sl uint8
	var h Header

	file, err := os.Open(filename)
	if err != nil {
		fmt.Print(err)
	}
	defer file.Close()

	err = binary.Read(file, binary.LittleEndian, &h)

	for {
		err = binary.Read(file, binary.LittleEndian, &id)
		if err != nil && err == io.EOF {
			break
		}

		err = binary.Read(file, binary.LittleEndian, &sl)
		if err != nil && err == io.EOF {
			break
		}

		instrument := make([]byte, sl)
		err = binary.Read(file, binary.LittleEndian, &instrument)
		if err != nil && err == io.EOF {
			break
		}

		beats := make([]byte, 16)
		err = binary.Read(file, binary.LittleEndian, &beats)
		if err != nil && err == io.EOF {
			break
		}

		t := track{int(id), string(instrument), boolString(beats)}
		tracks = append(tracks, t)

	}
	return &h, &tracks
}

func DecodeFile(path string) (string, error) {

	h, tracks := decodeDrum(path)
	var st string
	
	k := fmt.Sprint("Saved with HW Version: ", strings.Trim(string(h.Hwversion[0:]), "\x00"), "\n", "Tempo: ", h.Tempo, "\n")

	for _, t := range *tracks {
		st += fmt.Sprintf("(%d) %s\t%s\n", t.id, t.name, t.steps)
	}

	return k + st, nil
}
