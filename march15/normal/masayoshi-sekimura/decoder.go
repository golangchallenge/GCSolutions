package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

const spliceHeader = "SPLICE"

// Track is the low level representation of the
// each track contained in a .splice file.
type Track struct {
	ID    uint32
	Name  string
	Steps []byte
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version string
	Tempo   float32
	Tracks  []Track
}

func (p *Pattern) String() string {
	var b bytes.Buffer
	b.Write([]byte(fmt.Sprintf("Saved with HW Version: %s\n", p.Version)))
	b.Write([]byte(fmt.Sprintf("Tempo: %v\n", p.Tempo)))
	for _, t := range p.Tracks {
		b.Write([]byte(fmt.Sprintf("(%v) %v\t", t.ID, t.Name)))
		for i, s := range t.Steps {
			if i%4 == 0 {
				b.Write([]byte("|"))
			}
			if s == '\x00' {
				b.Write([]byte("-"))
			} else {
				b.Write([]byte("x"))
			}
		}
		b.Write([]byte("|\n"))
	}
	return b.String()
}

// Decoding stage.
// Header, Data Length, Version, Tempo and Tracks must appear in the order.
const (
	dsSeenHeader = iota
	dsSeenDataLength
	dsSeenVersion
	dsSeenTempo
	dsSeenTracks
)

type decoder struct {
	r          io.Reader
	p          Pattern
	dataLength uint8
	stage      int
	tmp        [256]byte
}

// A FormatError reports that the input is not a valid Splice file
type FormatError string

func (e FormatError) Error() string { return "drum: invalid format: " + string(e) }

func (d *decoder) checkHeader() error {
	_, err := io.ReadFull(d.r, d.tmp[:len(spliceHeader)])
	if err != nil {
		return err
	}
	if string(d.tmp[:len(spliceHeader)]) != spliceHeader {
		return FormatError("not a Splice file")
	}
	d.stage = dsSeenHeader
	return nil
}

func (d *decoder) parse() error {
	switch d.stage {
	case dsSeenHeader:
		if err := d.parseDataLength(); err != nil {
			return err
		}
		d.stage = dsSeenDataLength
	case dsSeenDataLength:
		if err := d.parseVersion(); err != nil {
			return err
		}
		d.stage = dsSeenVersion
	case dsSeenVersion:
		if err := d.parseTempo(); err != nil {
			return err
		}
		d.stage = dsSeenTempo
	case dsSeenTempo:
		if err := d.parseTracks(); err != nil {
			return err
		}
		d.stage = dsSeenTracks
	}
	return nil
}

func (d *decoder) parseDataLength() error {
	_, err := io.ReadFull(d.r, d.tmp[:8])
	if err != nil {
		return err
	}
	d.dataLength = uint8(d.tmp[7])
	return nil
}

func (d *decoder) parseVersion() error {
	n, err := io.ReadFull(d.r, d.tmp[:32])
	if err != nil {
		return err
	}
	d.dataLength -= uint8(n)

	i := bytes.Index(d.tmp[:32], []byte{0x00})
	d.p.Version = string(d.tmp[:i])
	return nil
}

func (d *decoder) parseTempo() error {
	n, err := io.ReadFull(d.r, d.tmp[:4])
	if err != nil {
		return err
	}
	d.dataLength -= uint8(n)

	var tempo float32
	buf := bytes.NewReader(d.tmp[:4])
	err = binary.Read(buf, binary.LittleEndian, &tempo)
	if err != nil {
		fmt.Println("binary.Read failed", err)
	}
	d.p.Tempo = tempo
	return nil
}

func (d *decoder) parseTracks() error {
	_, err := io.ReadFull(d.r, d.tmp[:d.dataLength])
	if err != nil {
		return err
	}

	i := 0
	var tracks []Track
	for {
		if i >= int(d.dataLength) {
			break
		}

		var id uint32
		buf := bytes.NewReader(d.tmp[i:(i + 4)])
		err = binary.Read(buf, binary.LittleEndian, &id)
		i += 4

		nameLength := int(d.tmp[i])
		i++

		name := string(d.tmp[i:(i + nameLength)])
		i += nameLength

		steps := d.tmp[i : i+16]
		i += 16

		tracks = append(tracks, Track{
			ID:    id,
			Name:  name,
			Steps: steps,
		})
	}
	d.p.Tracks = tracks

	return nil
}

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	d := &decoder{r: f}
	if err := d.checkHeader(); err != nil {
		return nil, err
	}
	for d.stage != dsSeenTracks {
		if err := d.parse(); err != nil {
			return nil, err
		}
	}
	return &d.p, nil
}
