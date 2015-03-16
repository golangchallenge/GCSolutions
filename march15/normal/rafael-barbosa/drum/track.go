package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// Track each patter has an array of tracks
// Each track has an array of steps
type Track struct {
	ID       [4]byte
	NameSize uint8
	Name     []byte
	Steps    [16]byte
}

// Decode decodes the track inside the pattern
func (t *Track) Decode(r io.Reader) error {
	if err := binary.Read(r, binary.LittleEndian, &t.ID); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &t.NameSize); err != nil {
		return err
	}

	if t.NameSize > maxDrumNameSize {
		return errors.New("Invalid Track")
	}

	t.Name = make([]byte, t.NameSize)

	if err := binary.Read(r, binary.LittleEndian, &t.Name); err != nil {
		return err
	}

	if err := binary.Read(r, binary.LittleEndian, &t.Steps); err != nil {
		return err
	}

	return nil
}

// String converts a Track to a string
func (t Track) String() string {
	return fmt.Sprintf("(%d) %s\t%s", binary.LittleEndian.Uint32(t.ID[:]), t.Name, stepsToCheetMusic(t.Steps[:]))
}

func stepsToCheetMusic(steps []byte) string {
	var buffer bytes.Buffer
	for index, step := range steps {
		if index%4 == 0 {
			buffer.WriteString("|")
		}
		if step == 1 {
			buffer.WriteString("x")
		} else {
			buffer.WriteString("-")
		}
	}
	buffer.WriteString("|")
	return buffer.String()
}
