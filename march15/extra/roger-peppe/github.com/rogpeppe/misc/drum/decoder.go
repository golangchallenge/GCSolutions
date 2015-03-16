// Package drum implements the encoding and decoding of .splice drum
// machine files. See http://golang-challenge.com/go-challenge1/ for
// more information
//
// See the drummachine package for a way of playing the drum machine
// patterns.
package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// NumBeats holds the number of beats in a drum machine track.
const NumBeats = 16

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	// Version holds the version of the splice file.
	Version string

	// Tempo holds the tempo of the pattern, in beats per minute.
	Tempo float32

	// Track holds the drum pattern tracks.
	Tracks []Track
}

// Track represents one track of the drum machine.
type Track struct {
	// Channel holds the numeric channel identifier.
	Channel int

	// Name holds the name of the channel.
	Name string

	// Beats holds all the beats in the track.
	// The value at Beats[t % NumBeats] specifies whether a beat is made at time t.
	Beats [NumBeats]bool
}

// DecodeFile decodes the drum machine pattern found at the provided
// file path.
func DecodeFile(path string) (*Pattern, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Decode(f)
}

// The binary format is as follows:
//
// signature: "SPLICE" 13 bytes, zero padded
// unknown: (0xc5 || 0x8f || 0x93 || 0x57)
// version: (string, 32 bytes, zero padded)
// tempo: (float32, little endian), 4 bytes
// tracks n * [
//	channel: (int32, little-endian) * 4 bytes
//	trackname: (namelen(int8), string[len])
//	beats: (0x0 | 0x1) * 4 * 4
// ]
//
// When the version is 0.708-alpha, we can
// have an apparently spurious SPLICE header
// at the end of the tracks.

const signature = "SPLICE"

// spuriousChannelVal holds the value that a track's channel
// will take when there's actually a "SPLICE" header
// there instead.
var spuriousChannelVal = int32(binary.LittleEndian.Uint32([]byte(signature[0:4])))

type header struct {
	Sig     [13]byte
	Unknown byte
	Version [32]byte
	Tempo   float32
}

type chanHeader struct {
	Channel int32
	NameLen byte
}

// Decode decodes the drum machine pattern read
// from the given reader.
func Decode(r io.Reader) (*Pattern, error) {
	var h header
	if err := binary.Read(r, binary.LittleEndian, &h); err != nil {
		return nil, fmt.Errorf("cannot read header: %v", err)
	}
	if sig := unpad(h.Sig[:]); sig != signature {
		return nil, fmt.Errorf("unexpected header, got %q, want %q", sig, signature)
	}
	var p Pattern
	p.Version = unpad(h.Version[:])
	if p.Version == "" {
		return nil, fmt.Errorf("no version found")
	}
	p.Tempo = h.Tempo
	for {
		var chanh chanHeader
		if err := binary.Read(r, binary.LittleEndian, &chanh); err != nil {
			if err == io.EOF {
				return &p, nil
			}
			return nil, fmt.Errorf("cannot read channel header: %v", err)
		}
		var t Track
		t.Channel = int(chanh.Channel)
		if p.Version == "0.708-alpha" && chanh.Channel == spuriousChannelVal && chanh.NameLen == 'C' {
			// It appears that 0.708 can have spurious extra header
			// and data tacked onto the end. We just ignore it for now.
			return &p, nil
		}
		name := make([]byte, chanh.NameLen)
		_, err := io.ReadFull(r, name)
		if err != nil {
			return nil, fmt.Errorf("cannot read channel name, size %d: %v", chanh.NameLen, err)
		}
		t.Name = string(name)
		var beats [NumBeats]byte
		_, err = io.ReadFull(r, beats[:])
		if err != nil {
			return nil, fmt.Errorf("cannot read channel beats: %v", err)
		}
		for i, beat := range beats {
			if beat != 0 && beat != 1 {
				return nil, fmt.Errorf("unexpected beat value %d in channel %q", beat, t.Name)
			}
			t.Beats[i] = beat != 0
		}
		p.Tracks = append(p.Tracks, t)
	}
}

func (p *Pattern) String() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "Saved with HW Version: %s\n", p.Version)
	fmt.Fprintf(&buf, "Tempo: %g\n", p.Tempo)

	for _, t := range p.Tracks {
		fmt.Fprintf(&buf, "(%d) %s\t", t.Channel, t.Name)
		writeBeats(&buf, t.Beats[:], 4)
		buf.WriteByte('\n')
	}
	return buf.String()
}

const defaultVersion = "1.0"

// MarshalBinary implements encoding.BinaryMarshaler
// for a pattern. The encoding it produces is readable
// by Decode and DecodeFile.
//
// The Version field in the pattern may be empty,
// in which case version 1.0 will be used.
func (p *Pattern) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer
	var h header
	copy(h.Sig[:], signature)
	h.Unknown = 0x57
	if p.Version == "" {
		copy(h.Version[:], defaultVersion)
	} else {
		copy(h.Version[:], p.Version)
	}
	h.Tempo = p.Tempo
	binary.Write(&buf, binary.LittleEndian, h)
	for _, t := range p.Tracks {
		if len(t.Name) > 255 {
			return nil, fmt.Errorf("track %d has name too long (%q)", t.Channel, t.Name)
		}
		binary.Write(&buf, binary.LittleEndian, chanHeader{
			Channel: int32(t.Channel),
			NameLen: byte(len(t.Name)),
		})
		buf.Write([]byte(t.Name))
		for _, b := range t.Beats {
			if b {
				buf.WriteByte(1)
			} else {
				buf.WriteByte(0)
			}
		}
	}
	return buf.Bytes(), nil
}

// writeBeats writes the beats in a track in |--x-| format.
// barLength holds the number of beats in a bar.
// The given beats must be a multiple of barLength.
func writeBeats(w *bytes.Buffer, beats []bool, barLength int) {
	w.WriteByte('|')
	for i := 0; i < len(beats); i += barLength {
		for j := 0; j < barLength; j++ {
			if beats[i+j] {
				w.WriteByte('x')
			} else {
				w.WriteByte('-')
			}
		}
		w.WriteByte('|')
	}
}

// unpad strips any trailing zero padding from
// the given byte string.
func unpad(s []byte) string {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] != 0 {
			return string(s[0 : i+1])
		}
	}
	return ""
}
