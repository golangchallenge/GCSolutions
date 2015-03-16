package drum

import (
	"encoding/binary"
	"fmt"
	"io"
)

// Track is a track with a 16 beat loop.
type Track struct {
	ID    int32
	Name  string
	Steps Steps
}

// NewTrack creates a new Track instance.
func NewTrack() *Track {
	return &Track{
		Steps: Steps{},
	}
}

func (t Track) String() string {
	return fmt.Sprintf("(%d) %s", t.ID, t.Name)
}

// Encode encodes the track in binary format to the writer.
func (t Track) Encode(w io.Writer) error {
	if err := binary.Write(w, binary.LittleEndian, t.ID); err != nil {
		return fmt.Errorf("unable to write ID, %v", err)
	}

	nLen := byte(len(t.Name))
	if err := binary.Write(w, binary.LittleEndian, nLen); err != nil {
		return fmt.Errorf("unable to write name length, %v", err)
	}

	if _, err := w.Write([]byte(t.Name)); err != nil {
		return fmt.Errorf("unable to write name, %v", err)
	}

	if err := t.Steps.Encode(w); err != nil {
		return fmt.Errorf("unable to encode steps, %v", err)
	}
	return nil
}

// DecodeTrack decodes the binary track, including the steps.
func DecodeTrack(r io.Reader) (t *Track, err error) {
	t = NewTrack()

	if binary.Read(r, binary.LittleEndian, &t.ID); err != nil {
		if err != io.EOF {
			err = fmt.Errorf("unable to read ID, %v", err)
		}
		return
	}

	var nLen byte
	if binary.Read(r, binary.LittleEndian, &nLen); err != nil {
		if err != io.EOF {
			err = fmt.Errorf("unable to read name length, %v", err)
		}
		return
	}

	p := make([]byte, nLen)
	if _, err = r.Read(p); err != nil {
		if err != io.EOF {
			err = fmt.Errorf("unable to read name, %v", err)
		}
		return
	}
	t.Name = string(p)

	t.Steps, err = DecodeSteps(r)
	return
}
