// Package drum is supposed to implement the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	// Bytes length for binary-encoded data header, contains "SPLICE" bytes and
	// padding bytes
	lenHeader = 13

	// Version string of Pattern version and padding
	lenPatternVersion = 32

	// Bytes length to store tempo (float32).
	lenPatternTempo = 4

	// Bytes length to store pattern header (version + tempo).
	lenPatternHeader = lenPatternVersion + lenPatternTempo

	// Bytes length to store track ID.
	lenTrackID = 4

	// Bytes length to store track hits.
	lenTrackHits = 16

	// Byte 13th of binary-encdoed data contains size of allocated bytes for
	// Pattern. The position counts from the first byte of the binary.
	posPatternSize = 13

	// Start position of Tracks bytes relative to the start bytes of Pattern
	// data.
	posPatternTracks = lenPatternVersion + lenPatternTempo

	// Position of byte defining track's name length, relative to the start
	// bytes of Pattern.
	posPatternTrackNameLen = 4

	// Start position of bytes containing track's name, relative to the start
	// bytes of Pattern.
	posPatternTrackName = 5
)

// Pattern is the high level representation of the drum pattern.
type Pattern struct {
	// Size retrieved from buf[pospatternSize]. This is the number of bytes
	// used to store Pattern information in binary-encoded data. This should
	// be set by Decoder.
	//
	// When Marshaling the Pattern into binary, this is not used at all. Instead
	// the size of allocated Pattern is retrieved from CalcSize().
	size uint8

	// Pattern version information. Padding bytes are removed during UnmarshalBinary.
	Version string

	// Pattern tempo.
	Tempo float32

	// Pattern tracks.
	Tracks []*Track
}

// MarshalBinary implements the encoding.BinaryMarshaler.
func (p *Pattern) MarshalBinary() (data []byte, err error) {
	buf := new(bytes.Buffer)

	// The header for binary-encoded data.
	if err = binary.Write(buf, binary.LittleEndian, [lenHeader]byte{'S', 'P', 'L', 'I', 'C', 'E'}); err != nil {
		return
	}

	// Bytes size for p.
	if err = binary.Write(buf, binary.LittleEndian, uint8(p.CalcSize())); err != nil {
		return
	}

	// Version.
	if err = binary.Write(buf, binary.LittleEndian, []byte(p.Version)); err != nil {
		return
	}

	// Padding for version.
	pad := make([]byte, lenPatternVersion-len(p.Version))
	if err = binary.Write(buf, binary.LittleEndian, pad); err != nil {
		return
	}

	// Tempo.
	if err = binary.Write(buf, binary.LittleEndian, p.Tempo); err != nil {
		return
	}

	// Tracks.
	for _, t := range p.Tracks {
		var tbin []byte
		if tbin, err = t.MarshalBinary(); err != nil {
			return
		}
		if err = binary.Write(buf, binary.LittleEndian, tbin); err != nil {
			return
		}
	}

	data = buf.Bytes()
	return
}

// CalcSize calculates number of bytes will be used by Pattern.
func (p *Pattern) CalcSize() uint8 {
	var s uint8 = 36
	for _, t := range p.Tracks {
		s += uint8(5 + len(t.Name) + lenTrackHits)
	}

	return s
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (p *Pattern) UnmarshalBinary(buf []byte) error {

	// Version
	p.Version = string(bytes.Replace(buf[:lenPatternVersion], []byte{0x00}, nil, -1))

	// Tempo
	if err := binary.Read(bytes.NewReader(buf[lenPatternVersion:lenPatternVersion+lenPatternTempo]), binary.LittleEndian, &p.Tempo); err != nil {
		return err
	}

	// Tracks.
	//
	// i represents cursor that advances to the first byte of a track. The first
	// position of a track is posPatternTracks and increased by the length of
	// a track:
	//
	//    lenTrackID + 1 + int(buf[i+4]) + lenTrackHits
	//
	// where 1 is byte to store track's name length and buf[i+4] is the actual
	// track's name length.
	//
	// m represents boundary position of tracks bytes. Since number of bytes of
	// p is known from p.size, allocated bytes for p.Tracks  would be p.size - 36,
	// where 36 is lenPatternVersion + lenPatternTempo.
	i := posPatternTracks
	m := int(posPatternTracks + (p.size - (lenPatternVersion + lenPatternTempo)))
	for {
		// Make sure we are not beyond Tracks bytes.
		if i >= m {
			break
		}

		// Parse track.
		track := new(Track)
		if err := track.UnmarshalBinary(buf[i : i+21+int(buf[i+4])]); err != nil {
			return err
		}

		// Append track.
		p.Tracks = append(p.Tracks, track)

		// Advances to next track.
		i += lenTrackID + 1 + int(buf[i+4]) + lenTrackHits
	}

	return nil
}

// String representation of a pattern p.
func (p *Pattern) String() string {
	return fmt.Sprintf("Saved with HW Version: %s\nTempo: %g\n%s", p.Version, p.Tempo, p.TracksString())
}

// TracksString repesents string of Pattern.Tracks.
func (p *Pattern) TracksString() string {
	var s string
	for _, t := range p.Tracks {
		s += t.String()
	}
	return s
}

// Track represents a single track in a pattern.
type Track struct {
	ID    uint32
	Name  string
	Steps Steps
}

// MarshalBinary implements the binary.BinaryMarshaler.
func (t *Track) MarshalBinary() (data []byte, err error) {
	buf := new(bytes.Buffer)

	// Track ID.
	if err = binary.Write(buf, binary.LittleEndian, t.ID); err != nil {
		return
	}

	// Track Name's length.
	if err = binary.Write(buf, binary.LittleEndian, uint8(len(t.Name))); err != nil {
		return
	}

	// Track Name.
	if err = binary.Write(buf, binary.LittleEndian, []byte(t.Name)); err != nil {
		return
	}

	// Track Steps.
	if err = binary.Write(buf, binary.LittleEndian, t.Steps); err != nil {
		return
	}

	data = buf.Bytes()
	return
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (t *Track) UnmarshalBinary(buf []byte) error {
	// Track ID.
	t.ID = binary.LittleEndian.Uint32(buf[:lenTrackID])

	// Track Name.
	n := make([]byte, buf[posPatternTrackNameLen])
	if err := binary.Read(bytes.NewReader(buf[posPatternTrackName:posPatternTrackName+buf[4]]), binary.LittleEndian, n); err != nil {
		return err
	}
	t.Name = string(n)

	// Track Steps.
	copy(t.Steps[:], buf[5+buf[4]:])

	return nil
}

// String representation of a track in a pattern.
func (t *Track) String() string {
	return fmt.Sprintf("(%d) %s\t%s\n", t.ID, t.Name, t.Steps)
}

// Steps represents 16 steps in a pattern. A byte 0x01 means a sound being
// triggered, while 0x00 means no sound being played for that step. In string,
// 0x00 is represented with '-' and 0x01 is represented with 'x'.
type Steps [16]byte

// String representation of steps.
func (s Steps) String() string {
	var ss string
	for i, b := range s {
		if i%4 == 0 {
			ss += "|"
		}

		switch b {
		case 0x00:
			ss += "-"
		case 0x01:
			ss += "x"
		}
	}
	ss += "|"

	return ss
}
