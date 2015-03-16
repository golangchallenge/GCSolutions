package drum

import (
	"encoding/binary"
	"io"
	"os"
)

// EncodeFile encodes the current Pattern into a drum machine file at the
// provided path.
func (p *Pattern) EncodeFile(path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	err = p.write(f)
	f.Close()
	return err
}

func (p *Pattern) write(w io.Writer) error {
	// 6 byte header magic of "SPLICE"
	buf := []byte("SPLICE")
	err := binary.Write(w, binary.LittleEndian, buf)
	if err != nil {
		return err
	}

	// 8 byte int length of remaining bytes in file (big-endian)
	var sz int64 = 36
	for _, part := range p.Parts {
		sz += 21 + int64(len(part.Name))
	}
	err = binary.Write(w, binary.BigEndian, sz)
	if err != nil {
		return err
	}

	// Version string is always 32 bytes
	buf = make([]byte, 32)
	for i, c := range p.HWVersion {
		buf[i] = byte(c)
	}
	err = binary.Write(w, binary.LittleEndian, buf)
	if err != nil {
		return err
	}

	// Tempo is a 4-byte float
	err = binary.Write(w, binary.LittleEndian, p.Tempo)
	if err != nil {
		return err
	}

	for _, part := range p.Parts {
		err = part.write(w)
		if err != nil {
			break
		}
	}

	return err
}

func (p *Part) write(w io.Writer) error {
	err := binary.Write(w, binary.LittleEndian, p.ID)
	if err != nil {
		return err
	}

	nameLen := uint8(len(p.Name))
	err = binary.Write(w, binary.LittleEndian, nameLen)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, []byte(p.Name))
	if err != nil {
		return err
	}

	// the on-disk format isn't very compact, one whole byte per step even
	// though only 1 bit is necessary.
	stepBuf := make([]byte, 16)
	b := uint16(1) << 15
	for i := 0; i < 16; i++ {
		if (p.Steps & b) != 0 {
			stepBuf[i] = 1
		}
		b >>= 1
	}
	return binary.Write(w, binary.LittleEndian, stepBuf)
}
