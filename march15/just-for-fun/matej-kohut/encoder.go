package drum

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"os"
)

// EncodeToFile writes pattern to a file provided by path
func (p *Pattern) EncodeToFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	err = p.Write(writer)
	if err != nil {
		return err
	}
	writer.Flush()
	return nil
}

func (p *Pattern) Write(buf io.Writer) error {
	data := &bytes.Buffer{}
	versionData := []byte(p.version)
	if len(versionData) > 32 {
		versionData = versionData[:32]
	}
	if len(versionData) < 32 {
		versionData = append(versionData, make([]byte, 32-len(versionData))...)
	}
	_, err := data.Write(versionData)
	if err != nil {
		return err
	}

	err = binary.Write(data, binary.LittleEndian, p.Tempo)
	if err != nil {
		return err
	}

	for _, instrument := range p.Instruments {
		err := instrument.Write(data)
		if err != nil {
			return err
		}
	}

	headerData := &struct {
		Head     [6]byte
		DataSize uint64
	}{[6]byte{'S', 'P', 'L', 'I', 'C', 'E'}, uint64(data.Len())}
	err = binary.Write(buf, binary.BigEndian, headerData)
	if err != nil {
		return err
	}

	_, err = buf.Write(data.Bytes())
	if err != nil {
		return err
	}

	_, err = buf.Write(p.footer)
	if err != nil {
		return err
	}
	return nil
}

func (i *Instrument) Write(buf io.Writer) error {
	err := binary.Write(buf, binary.BigEndian, i.ID)
	if err != nil {
		return err
	}

	nameLength := uint32(len(i.Name))
	err = binary.Write(buf, binary.BigEndian, nameLength)
	if err != nil {
		return err
	}

	_, err = buf.Write([]byte(i.Name))
	if err != nil {
		return err
	}

	err = binary.Write(buf, binary.BigEndian, i.Steps)
	if err != nil {
		return err
	}

	return nil

}
