package drum

import (
	"bytes"
	"encoding/binary"
	"io/ioutil"
)

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version string
	Tempo   float32
	Tracks  []Track
}

type Track struct {
	ID    byte
	Name  string
	Steps []byte
}

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	p := &Pattern{
		Version: decodeVersion(data),
		Tempo:   decodeTempo(data),
		Tracks:  decodeTracks(data),
	}

	return p, nil
}

func decodeVersion(d []byte) string {
	v := d[14:46]
	i := 31
	// shave off \x00 values from rear of string
	for ; i >= 0; i-- {
		if v[i] != '\x00' {
			break
		}
	}

	return string(v[0 : i+1])
}

func decodeTempo(d []byte) float32 {
	tData := d[46:50]
	var t float32
	buf := bytes.NewReader(tData)
	err := binary.Read(buf, binary.LittleEndian, &t)
	if err != nil {
		panic(err)
	}

	return t
}

func decodeTracks(d []byte) []Track {
	tLen := int(d[13])

	ts := []Track{}
	for i := 50; i <= tLen; {
		tID := d[i]
		i += 4

		tNameLen := int(d[i])
		i++
		tName := string(d[i : i+tNameLen])
		i += tNameLen

		s := d[i : i+16]
		i += 16

		t := Track{
			ID:    tID,
			Name:  tName,
			Steps: s,
		}

		ts = append(ts, t)
	}

	return ts
}
