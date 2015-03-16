// Package drum is supposed to implement the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"strings"
)

const (
	headerSize     = 14
	versionSize    = 32
	tempoSize      = 4
	idSize         = 4
	gridSize       = 16
	nameLengthSize = 1
)

func checkForSize(content []byte, size int, element string) error {
	if len(content) < size {
		return errors.New("Not enough data to read " + element)
	}
	return nil
}
func (p *Pattern) Decode(content []byte) error {
	bytePointer := 0

	// read header
	if err := checkForSize(content, bytePointer+headerSize, "header"); err != nil {
		return err
	}
	p.Header = strings.TrimRight(string(content[0:headerSize]), "\x00")

	// read version
	bytePointer = bytePointer + headerSize
	if err := checkForSize(content, bytePointer+versionSize, "version"); err != nil {
		return err
	}
	p.Version = strings.TrimRight(string(content[bytePointer:bytePointer+versionSize]), "\x00")

	// read tempo
	bytePointer = bytePointer + versionSize
	if err := checkForSize(content, bytePointer+tempoSize, "tempo"); err != nil {
		return err
	}
	buf := bytes.NewReader(content[bytePointer : bytePointer+tempoSize])
	if err := binary.Read(buf, binary.LittleEndian, &p.Tempo); err != nil {
		return err
	}

	// read tracks
	bytePointer = bytePointer + tempoSize
	trackNumber := 0
	p.Tracks = make([]*Track, 0)
	for bytePointer <= len(content) {
		if newBytePointer, err := p.decodeTrack(content, trackNumber, bytePointer); err != nil {
			return err
		} else {
			bytePointer = newBytePointer
		}
		trackNumber = trackNumber + 1
	}
	return nil
}

func (p *Pattern) decodeTrack(content []byte, trackNumber, bytePointer int) (int, error) {
	t := &Track{}
	// read Track Id
	if err := checkForSize(content, bytePointer+idSize, "track ID"); err != nil {
		return 0, err
	}
	buf := bytes.NewReader(content[bytePointer : bytePointer+idSize])
	if err := binary.Read(buf, binary.LittleEndian, &t.Id); err != nil {
		return 0, err
	}

	// read name length
	bytePointer = bytePointer + idSize
	if err := checkForSize(content, bytePointer+nameLengthSize, "track name"); err != nil {
		return 0, err
	}
	nameLength := int(content[bytePointer])

	// read track name
	bytePointer = bytePointer + nameLengthSize
	if err := checkForSize(content, bytePointer+nameLength, "track name"); err != nil {
		return 0, err
	}
	t.Name = string(content[bytePointer : bytePointer+nameLength])

	// read grid
	bytePointer = bytePointer + nameLength
	if err := checkForSize(content, bytePointer+gridSize, "track beat grid"); err != nil {
		return 0, err
	}

	if err := t.readBeatGrid(content, bytePointer); err != nil {
		return 0, err
	}
	bytePointer = bytePointer + gridSize

	p.Tracks = append(p.Tracks, t)

	return bytePointer, nil
}

func (t *Track) readBeatGrid(content []byte, bytePointer int) error {
	beat := "|"
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			if int(content[bytePointer+(4*i)+j]) == 0 {
				beat = beat + "-"
			} else if int(content[bytePointer+(4*i)+j]) == 1 {
				beat = beat + "x"
			} else {
				errors.New("Illegal encoding for track beat grid")
			}
		}
		beat = beat + "|"
	}
	t.Steps = beat
	return nil
}
