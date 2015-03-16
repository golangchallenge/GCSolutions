package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"strings"
)

// ErrNoSpliceHeader is returned by Read, when the data are not proper
// .splice file data
var ErrNoSpliceHeader = errors.New("drum: file has not proper headers")

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	dataFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer dataFile.Close()

	p := &Pattern{}
	err = p.Read(dataFile)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// Reads data from buffer to drum pattern
func (p *Pattern) Read(buf io.Reader) error {
	headerData := &struct {
		Head     [6]byte
		DataSize uint64
	}{}
	err := binary.Read(buf, binary.BigEndian, headerData)
	if err != nil {
		return err
	}

	if headerData.Head != [6]byte{'S', 'P', 'L', 'I', 'C', 'E'} {
		return ErrNoSpliceHeader
	}

	if headerData.DataSize == 0 {
		return nil
	}

	data := make([]byte, headerData.DataSize)
	_, err = buf.Read(data)
	if err != nil {
		return err
	}
	dataStream := bytes.NewReader(data)

	var versionData [32]byte
	err = binary.Read(dataStream, binary.LittleEndian, &versionData)
	if err != nil {
		return err
	}
	p.version = strings.Trim(string(versionData[:]), string('\x00'))

	err = binary.Read(dataStream, binary.LittleEndian, &p.Tempo)
	if err != nil {
		return err
	}

	for {
		instrument := &Instrument{}
		err := instrument.Read(dataStream)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		p.Instruments = append(p.Instruments, instrument)
	}

	footerBuffer := bytes.Buffer{}
	_, err = footerBuffer.ReadFrom(buf)
	if err != nil {
		return err
	}
	p.footer = footerBuffer.Bytes()
	return nil
}

// Reads data from straem to drum pattern instrument
func (i *Instrument) Read(buf io.Reader) error {
	err := binary.Read(buf, binary.BigEndian, &i.ID)
	if err != nil {
		return err
	}
	var nameLength uint32
	err = binary.Read(buf, binary.BigEndian, &nameLength)
	if err != nil {
		return err
	}
	name := make([]byte, nameLength)
	_, err = buf.Read(name)
	if err != nil {
		return err
	}
	i.Name = string(name)

	err = binary.Read(buf, binary.BigEndian, &i.Steps)
	if err != nil {
		return err
	}
	return nil
}
