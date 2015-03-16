package drum

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	p := &Pattern{}

	r, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	splice := [6]byte{}
	if err := binary.Read(r, binary.BigEndian, &splice); err != nil {
		return nil, err
	}
	if splice != [6]byte{'S', 'P', 'L', 'I', 'C', 'E'} {
		return nil, fmt.Errorf("Header of the input file must start with 'SPLICE', but starts with: '%s'", string(splice[:]))
	}

	var bodyLength uint64
	if err := binary.Read(r, binary.BigEndian, &bodyLength); err != nil {
		return nil, err
	}

	// we limit the amount of bytes to read according to length declared in the header
	lr := io.LimitReader(r, int64(bodyLength))

	hwVersion := [32]byte{}
	if err := binary.Read(lr, binary.BigEndian, &hwVersion); err != nil {
		return nil, err
	}
	p.HwVersion = string(bytes.Trim(hwVersion[:], "\x00"))

	if err := binary.Read(lr, binary.LittleEndian, &p.Tempo); err != nil {
		return nil, err
	}

	// now we load all the instruments
	for {
		if ins, err := decodeInstrument(lr); err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, fmt.Errorf("Error loading instrument: %v", err)
			}
		} else {
			p.Instruments = append(p.Instruments, ins)
		}
	}

	return p, nil
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	HwVersion   string
	Tempo       float32
	Instruments []*Instrument
}

// Instrument is a representation of an instrument
// along with its name and pattern
type Instrument struct {
	ID      byte
	Label   string
	Pattern [16]bool
}

func (p Pattern) String() string {
	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	fmt.Fprintf(w, "Saved with HW Version: %s\n", p.HwVersion)
	fmt.Fprintf(w, "Tempo: %v\n", p.Tempo)

	for _, in := range p.Instruments {
		fmt.Fprintf(w, "%s\n", in)
	}

	w.Flush()
	return b.String()
}

func (in Instrument) String() string {
	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	// print id and label
	fmt.Fprintf(w, "(%d) %s\t", in.ID, in.Label)

	// now we print the pattern
	for i := 0; i < 4; i++ {
		fmt.Fprintf(w, "|")
		for j := 0; j < 4; j++ {
			if in.Pattern[i*4+j] {
				fmt.Fprintf(w, "x")
			} else {
				fmt.Fprintf(w, "-")
			}
		}
	}
	fmt.Fprintf(w, "|")

	w.Flush()
	return b.String()
}

// decode one instrument record from the binary file
func decodeInstrument(r io.Reader) (*Instrument, error) {
	ins := &Instrument{}

	if err := binary.Read(r, binary.BigEndian, &ins.ID); err != nil {
		return nil, err
	}

	var labelLen uint32
	if err := binary.Read(r, binary.BigEndian, &labelLen); err != nil {
		return nil, err
	}

	label := make([]byte, labelLen)
	if err := binary.Read(r, binary.BigEndian, &label); err != nil {
		return nil, err
	}
	ins.Label = string(label[:])

	pattern := [16]byte{}
	if err := binary.Read(r, binary.BigEndian, &pattern); err != nil {
		return nil, err
	}

	// convert [16]byte to [16]bool
	for i, p := range pattern {
		if p == 0 {
			ins.Pattern[i] = false
		} else {
			ins.Pattern[i] = true
		}
	}

	return ins, nil
}
