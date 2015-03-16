package drum

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	p := &Pattern{}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	err = p.read(f)
	f.Close()
	return p, err
}

func (p *Pattern) read(r io.Reader) error {
	// 6 byte header magic of "SPLICE"
	buf := make([]byte, 6)
	err := binary.Read(r, binary.LittleEndian, &buf)
	if err != nil {
		return err
	}
	if string(buf) != "SPLICE" {
		return fmt.Errorf("invalid format (no SPLICE header)")
	}

	// 8 byte int length of remaining bytes in file (big-endian)
	var sz int64
	err = binary.Read(r, binary.BigEndian, &sz)
	if err != nil {
		return err
	}
	if sz <= 36 {
		return fmt.Errorf("invalid data size (%d <= 36)", sz)
	}

	// Version string is always 32 bytes
	buf = make([]byte, 32)
	err = binary.Read(r, binary.LittleEndian, &buf)
	if err != nil {
		return err
	}
	// for some reason Go keeps the extra null bytes in the string rep, so
	// we have to trim them out.
	p.HWVersion = strings.Trim(string(buf), "\x00")

	// Tempo is a 4-byte float
	err = binary.Read(r, binary.LittleEndian, &p.Tempo)
	if err != nil {
		return err
	}

	sz -= 36 // Version + Tempo bytes
	for sz > 0 {
		part := &Part{}
		err = part.read(r)
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			break
		}
		p.Parts = append(p.Parts, part)

		// 21 = 16 byte steps + 4 byte ID + 1 byte nameLen
		sz -= 21 + int64(len(part.Name))
	}

	return err
}

func (p *Part) read(r io.Reader) error {
	err := binary.Read(r, binary.LittleEndian, &p.ID)
	if err != nil {
		return err
	}

	var nameLen uint8
	err = binary.Read(r, binary.LittleEndian, &nameLen)
	if err != nil {
		return err
	}
	if nameLen == 0 {
		return fmt.Errorf("invalid zero-length name")
	}
	nameBuf := make([]byte, nameLen)
	err = binary.Read(r, binary.LittleEndian, &nameBuf)
	if err != nil {
		return err
	}
	p.Name = string(nameBuf)

	// the on-disk format isn't very compact, one whole byte per step even
	// though only 1 bit is necessary.
	stepBuf := make([]byte, 16)
	err = binary.Read(r, binary.LittleEndian, &stepBuf)
	if err != nil {
		return err
	}
	b := uint16(1) << 15
	for i := 0; i < 16; i++ {
		if stepBuf[i] > 1 {
			return fmt.Errorf("unknown step type '%d' (should be 0 or 1)", stepBuf[i])
		}
		if stepBuf[i] == 1 {
			p.Steps |= b
		}
		b >>= 1
	}
	return nil
}
