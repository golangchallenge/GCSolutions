package drum

import (
	"encoding/binary"
	"io"
	"strings"
)

const version = "0.808-alpha"

func (p Pattern) writeHeader(w io.Writer) error {
	_, err := w.Write([]byte("SPLICE\x00\x00\x00\x00\x00\x00"))
	return err
}

func (p Pattern) writeSize(w io.Writer) error {
	var tracks uint16

	for _, t := range p.Tracks {
		tracks += uint16(5 + len(t.Name) + 16)
	}

	size := uint16(32 + 4 + tracks)
	return binary.Write(w, binary.BigEndian, size)
}

func (p Pattern) writeVersion(w io.Writer) error {
	v := version + strings.Repeat("\x00", 32-len(version))
	_, err := w.Write([]byte(v))
	return err
}

func (p Pattern) writeTempo(w io.Writer) error {
	return binary.Write(w, binary.LittleEndian, p.Tempo)
}

func (t Track) writeID(w io.Writer) error {
	return binary.Write(w, binary.BigEndian, t.ID)
}

func (t Track) writeName(w io.Writer) error {
	bytes := []byte(t.Name)
	err := binary.Write(w, binary.BigEndian, uint32(len(bytes)))
	_, err = w.Write(bytes)
	return err
}

func (t Track) writeSteps(w io.Writer) error {

	bytes := make([]byte, 16)
	for i := 0; i < 16; i++ {
		if t.Steps[i] {
			bytes[i] = 1
		}
	}

	_, err := w.Write(bytes)
	return err

}

func (t Track) write(w io.Writer) error {
	err := t.writeID(w)
	err = t.writeName(w)
	err = t.writeSteps(w)
	return err
}
