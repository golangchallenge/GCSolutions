// Package drum decodes .splice drum machine files.
package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
)

// Pattern represents a drum pattern of a .splice file.
type Pattern struct {
	HWVersion string  // readable string
	Tempo     float32 // beats per minute
	Tracks    []Track // variable number of tracks
}

// Track represents a single part or sound of a drum pattern.
type Track struct {
	ID    int32
	Name  string   // typically an instrument or sound name
	Steps [16]byte // use constants Rest and Beat
}

// Rest and Beat are possible values for steps of a track.
const (
	Rest = 0
	Beat = 1
)

// DecodeFile decodes the drum machine file found at the given path
// and returns a parsed pattern.
//
// Valid tracks following a valid header are returned without error.
//
// An error is returned if the file header cannot be read or decoded.
//
// If invalid track or non-track data is encountered decoding terminates
// and the pattern read so far is returned without error.
func DecodeFile(path string) (*Pattern, error) {
	d, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	b := bytes.NewBuffer(d)
	// File format interpretation here is reverse-engineered and so may be
	// incomplete or inaccurate.
	//
	// As decoded, a file has a fixed size file header followed by a variable
	// number of variable length track records.
	//
	// File header format, 50 bytes total:
	var hdr struct {
		FileID    [6]byte  // required to be "SPLICE"
		_         [8]byte  // unknown use, not decoded
		HWVersion [32]byte // human readable version string, null terminated
		Tempo     float32
	}
	if err = binary.Read(b, binary.LittleEndian, &hdr); err != nil {
		return nil, fmt.Errorf("error reading SPLICE file: %v", err)
	}
	if string(hdr.FileID[:]) != "SPLICE" {
		return nil, fmt.Errorf("SPLICE file ID not found")
	}
	// track format, variable size:
	//
	// 4 bytes, integer track id
	// 1 byte, (n) length of track name
	// n bytes, human readable track name
	// 16 bytes, steps, 1 step per byte, 0=rest, 1=beat
	var tHdr struct {
		ID  int32
		Len byte
	}
	var tracks []Track
	for b.Len() > 0 {
		err = binary.Read(b, binary.LittleEndian, &tHdr)
		if err != nil || b.Len() < int(tHdr.Len)+16 {
			break // short track, terminate decoding
		}
		tk := Track{
			ID:   tHdr.ID,
			Name: string(b.Next(int(tHdr.Len))),
		}
		copy(tk.Steps[:], b.Next(16))
		tracks = append(tracks, tk)
	}
	return &Pattern{zstr(hdr.HWVersion[:]), hdr.Tempo, tracks}, nil
}

// convert null teminated byte slice to string
func zstr(z []byte) string {
	l := bytes.IndexByte(z, 0)
	if l >= 0 {
		z = z[:l]
	}
	return string(z)
}

// String returns a multi-line human readable representation of a drum pattern.
func (p *Pattern) String() string {
	s := fmt.Sprintf("Saved with HW Version: %s\nTempo: %g\n",
		p.HWVersion, p.Tempo)
	var steps [16]byte
	for _, t := range p.Tracks {
		for i, b := range t.Steps {
			m := byte('-') // output non-beats as rests
			if b == Beat {
				m = 'x'
			}
			steps[i] = m
		}
		s += fmt.Sprintf("(%d) %s\t|%s|%s|%s|%s|\n",
			t.ID, t.Name, steps[:4], steps[4:8], steps[8:12], steps[12:])
	}
	return s
}
