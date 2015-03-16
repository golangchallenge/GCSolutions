package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

func readTrack(spliceFile io.Reader) *Track {

	// Now read the tracks
	// le-int32 id
	//     int8 name size
	//  byte[n] name, no null term
	// byte[16] 16 beats - 0/1 on/off

	var id int32
	var nameSize int8
	var rawName []byte
	var name string
	var beats = make([]byte, 16)

	binary.Read(spliceFile, binary.LittleEndian, &id)
	binary.Read(spliceFile, binary.LittleEndian, &nameSize)

	rawName = make([]byte, nameSize)
	spliceFile.Read(rawName)
	name = string(rawName)

	spliceFile.Read(beats)

	return &Track{id, name, beats}
}

func NewTrack(ID int32, name string) *Track {
	return &Track{ID: ID, Name: name, Beats: make([]byte, 16, 16)}
}

func (t *Track) String() string {
	var buffer bytes.Buffer

	fmt.Fprintf(&buffer, "(%d) %s\t", t.ID, t.Name)
	for j := 0; j < len(t.Beats); j++ {
		if j%4 == 0 {
			buffer.WriteString("|")
		}
		if t.Beats[j] == 0 {
			buffer.WriteString("-")
		} else {
			buffer.WriteString("x")
		}
	}
	buffer.WriteString("|")

	return buffer.String()
}

func (t *Track) On(idx ...int) {
	for i := 0; i < len(idx); i++ {
		iv := idx[i]
		if iv < 0 || iv > 15 {
			continue
		}
		t.Beats[iv] = 1
	}
}

func (t *Track) Off(idx ...int) {
	for i := 0; i < len(idx); i++ {
		iv := idx[i]
		if iv < 0 || iv > 15 {
			continue
		}
		t.Beats[iv] = 0
	}
}

func (t *Track) SetString(bin string) {
	for i, c := range bin[:min(16, len(bin))] {
		if c == 'x' || c == '1' || c == 'X' {
			t.Beats[i] = 1
		} else {
			t.Beats[i] = 0
		}
	}
}

// Set the value for a given beat
func (t *Track) Set(beat int, value bool) error {
	if beat < 0 || beat > 15 {
		return errors.New("Invalid beat, must be between [0,15]")
	}
	if value {
		t.Beats[beat] = 1
	} else {
		t.Beats[beat] = 0
	}

	return nil
}

func (t *Track) SetAll(values []byte, offset int) {
	if offset < 0 || offset > 15 {
		return
	}

	for j := offset; j < min(len(values), 16-offset); j++ {
		if values[j] != 0 {
			t.Beats[j] = 1
		} else {
			t.Beats[j] = 0
		}
	}
}

func (t *Track) SetStep(value bool, step int, offset int) {
	if offset < 0 || offset > 15 {
		return
	}

	for j := offset; j < 16; j += step {
		if value {
			t.Beats[j] = 1
		} else {
			t.Beats[j] = 0
		}
	}
}

func (t *Track) Clear() {
	for i := 0; i < len(t.Beats); i++ {
		t.Beats[i] = 0
	}
}

func (t *Track) Invert() {
	for i := 0; i < len(t.Beats); i++ {
		if t.Beats[i] == 0 {
			t.Beats[i] = 1
		} else {
			t.Beats[i] = 0
		}
	}
}

func (t *Track) Write(spliceFile io.Writer) (int, error) {
	var written int

	binary.Write(spliceFile, binary.LittleEndian, t.ID)
	written += 4
	binary.Write(spliceFile, binary.LittleEndian, int8(len(t.Name)))
	written++
	spliceFile.Write([]byte(t.Name))
	written += len(t.Name)
	spliceFile.Write(t.Beats)
	written += len(t.Beats)

	return written, nil
}
