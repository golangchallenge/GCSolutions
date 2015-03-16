package drum

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

const (
	nquarters = 4
	quarter   = 4
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	p := &Pattern{}

	file, _ := os.Open(path)
	buf := bufio.NewReader(file)

	header := make([]byte, 6)
	binary.Read(buf, binary.BigEndian, header)

	sizeBuf := make([]byte, 8)
	binary.Read(buf, binary.BigEndian, sizeBuf)
	size := binary.BigEndian.Uint64(sizeBuf)
	var n uint64

	version := make([]byte, 32)
	binary.Read(buf, binary.BigEndian, version)
	p.Version = strings.TrimRight(string(version), "\x00")
	n += 32

	binary.Read(buf, binary.LittleEndian, &p.Tempo)
	n += 4

	for n < size {
		var t Track

		binary.Read(buf, binary.LittleEndian, &t.ID)
		n += 4

		len, _ := binary.ReadUvarint(buf)
		n++

		instrument := make([]byte, len)
		binary.Read(buf, binary.LittleEndian, instrument)
		t.Instrument = string(instrument)
		n += uint64(len)

		for i := range t.Steps {
			b, _ := buf.ReadByte()
			t.Steps[i] = int8(b)
			n++
		}

		p.Tracks = append(p.Tracks, t)
	}

	return p, nil
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version string
	Tempo   float32
	Tracks  []Track
}

// Track is the representation of a single track
type Track struct {
	ID         int32
	Instrument string
	Steps      [nquarters * quarter]int8
}

func (p Pattern) String() string {
	buf := bytes.NewBufferString("")

	fmt.Fprintln(buf, "Saved with HW Version:", p.Version)
	fmt.Fprintln(buf, "Tempo:", p.Tempo)

	for _, t := range p.Tracks {
		fmt.Fprintln(buf, t)
	}

	return buf.String()
}

func (t Track) String() string {
	buf := bytes.NewBufferString("")

	note := map[int8]string{
		0: "-",
		1: "x",
	}

	fmt.Fprintf(buf, "(%v) %v\t", t.ID, t.Instrument)

	for i, s := range t.Steps {
		if i%quarter == 0 {
			fmt.Fprint(buf, "|")
		}

		fmt.Fprint(buf, note[s])
	}
	fmt.Fprint(buf, "|")

	return buf.String()
}
