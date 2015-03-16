package drum

import (
	"bytes"
	"encoding/binary"
	"os"
)

// EncodeFile writes the current drum pattern into the specified file
func (p *Pattern) EncodeFile(path string) error {
	// it would make more sense to pass in a io.Writer
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	buffer := &bytes.Buffer{}
	// create tracks
	for _, t := range p.Tracks {
		// TODO error checks!
		buffer.WriteByte(t.ID)
		binary.Write(buffer, binary.BigEndian, int32(len(t.Name)))
		buffer.WriteString(t.Name) // cast to []byte ?
		for _, s := range t.Steps {
			if s {
				buffer.WriteByte(1)
			} else {
				buffer.WriteByte(0)
			}
		}
	}

	// TODO use n
	_, err = file.Write(spliceMagic)
	if err != nil {
		return err
	}

	// write tracks len + version tring + tempo
	err = binary.Write(file, binary.BigEndian, int64(buffer.Len()+32+4))

	_, err = file.WriteString(p.Version)
	for i := len(p.Version); i < 32; i++ {
		_, err = file.Write([]byte{0x00})
	}

	err = binary.Write(file, binary.LittleEndian, p.Tempo)

	_, err = file.Write(buffer.Bytes())

	return nil
}
