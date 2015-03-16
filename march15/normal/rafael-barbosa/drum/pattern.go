package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"strings"
)

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Name    [14]byte
	Version [32]byte
	Tempo   [4]byte
	Tracks  []Track
}

// Decode decodes the file into a Pattern struct
func (p *Pattern) Decode(r io.Reader) error {
	if err := binary.Read(r, binary.LittleEndian, &p.Name); err != nil {
		return err
	}

	if err := binary.Read(r, binary.LittleEndian, &p.Version); err != nil {
		return err
	}

	if err := binary.Read(r, binary.LittleEndian, &p.Tempo); err != nil {
		return err
	}

	for true {
		t := Track{}
		err := t.Decode(r)
		if err != nil {
			break
		}
		p.Tracks = append(p.Tracks, t)
	}
	return nil
}

// String converts a Pattern to a string
func (p Pattern) String() string {
	return fmt.Sprintf(`Saved with HW Version: %s
Tempo: %g
%s
`, bytes.Trim(p.Version[:], "\x00"), float32frombytes(p.Tempo[:]), p.printTracks())
}

func float32frombytes(tempo []byte) float32 {
	bits := binary.LittleEndian.Uint32(tempo)
	float := math.Float32frombits(bits)
	return float
}

func (p Pattern) printTracks() string {
	var buffer []string
	for _, track := range p.Tracks {
		buffer = append(buffer, track.String())
	}
	return strings.Join(buffer, "\n")
}
