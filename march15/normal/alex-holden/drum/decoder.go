// The author disclaims copyright to this source code.  In place of
// a legal notice, here is a blessing:
//
//    May you do good and not evil.
//    May you find forgiveness for yourself and forgive others.
//    May you share freely, never taking more than you give.

package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
)

// Pattern Binary Format
//
//  Offset     Length    Contents
//     0       6  bytes  "SPLICE"
//     6       4  bytes  RESERVED
//    10       4  bytes  <File size(x): uint32, Big-endian>
//    14      32  bytes  <Version name>
//    46       4  bytes  <Tempo: float32, Little-endian>
//  [ 50      (s) bytes  <Tracks: (s) = (x) - 36> ]*
//
//  Track:
//     0       2  byte   <Track id: uint16, Little-endian>
//     2       2  bytes  RESERVED
//     4       1  byte   <Track name size(n): uint8>
//     5      (n) bytes  <Track name>
//     5+(n)  16  bytes  <Track data>

// Pattern is the high level representation of the drum pattern contained
// in a .splice file.
type Pattern struct {
	Version string
	Tempo   float32
	Tracks  []*Track
}

func (p Pattern) String() string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("Saved with HW Version: %s\nTempo: %v\n", p.Version, p.Tempo))
	for _, t := range p.Tracks {
		buf.WriteString(t.String())
	}
	return buf.String()
}

// A Track represents an individual instrument or sound within a Pattern
type Track struct {
	ID    uint16
	Name  string
	Beats [16]bool
}

func (t Track) String() string {
	var buf bytes.Buffer
	for i, b := range t.Beats {
		if b {
			buf.WriteRune('x')
		} else {
			buf.WriteRune('-')
		}
		if i%4 == 3 {
			buf.WriteRune('|')
		}
	}
	return fmt.Sprintf("(%v) %v\t|%v\n", t.ID, t.Name, buf.String())
}

// DecodeFile decodes the drum machine file found at the provided path and
// returns a pointer to a parsed Pattern which is the entry point to the data.
// The file is expected to conform to the following format.
func DecodeFile(fp string) (*Pattern, error) {
	b, err := ioutil.ReadFile(fp)
	if err != nil {
		return nil, fmt.Errorf("drum: %v", err)
	}
	p := &Pattern{}
	return p, p.UnmarshalBinary(b)
}

// EncodeFile takes a Pattern, encodes it to binary, and saves it to the
// specified file path
func (p *Pattern) EncodeFile(fp string) error {
	data, err := p.MarshalBinary()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(fp, data, 0644)
}

// MarshalBinary encodes a given Pattern to binary form.
func (p Pattern) MarshalBinary() ([]byte, error) {
	tds := 0
	for _, t := range p.Tracks {
		tds += len(t.Name) + 21
	}

	buf := make([]byte, tds+50)
	copy(buf[0:], "SPLICE")
	binary.BigEndian.PutUint32(buf[10:], uint32(tds+36))
	copy(buf[14:], p.Version)
	binary.LittleEndian.PutUint32(buf[46:], math.Float32bits(p.Tempo))

	ofs := 50
	for _, t := range p.Tracks {
		nl := len(t.Name)
		binary.LittleEndian.PutUint16(buf[ofs:], t.ID)
		buf[ofs+4] = uint8(nl)
		copy(buf[ofs+5:], t.Name)

		for i, b := range t.Beats {
			if b {
				buf[ofs+nl+5+i] = 1
			}
		}
		ofs += nl + 21
	}
	return buf, nil
}

// UnmarshalBinary marshals the given binary data into the Pattern receiver.
func (p *Pattern) UnmarshalBinary(data []byte) error {
	if len(data) < 50 {
		return errors.New("drum: file too small (<50 bytes)")
	} else if string(data[0:6]) != "SPLICE" {
		return errors.New("drum: missing 'SPLICE' declaration")
	}

	tds := int(binary.BigEndian.Uint32(data[10:14]) - 36)
	if len(data) < tds+50 {
		return errors.New("drum: file too small to hold declared track data")
	}

	p.Version = string(bytes.Trim(data[14:46], "\x00"))
	p.Tempo = math.Float32frombits(binary.LittleEndian.Uint32(data[46:50]))

	data = data[50 : 50+tds]

	// decode the tracks
	for i := 0; i < len(data); {
		n := int(data[i+4]) // track name size
		t := Track{}
		t.ID = binary.LittleEndian.Uint16(data[i : i+2])
		t.Name = string(data[i+5 : i+5+n])

		for ofs, j := i+5+n, 0; j < 16; j++ {
			if data[ofs+j] == 1 {
				t.Beats[j] = true
			}
		}

		p.Tracks = append(p.Tracks, &t)
		i += 21 + n
	}
	return nil
}
