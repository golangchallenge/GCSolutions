package drum

import (
	"encoding/hex"
	"errors"
	"strconv"
)

// A PatternTracksParser is responsible for parsing the musical
// section of a pattern file.
type PatternTracksParser struct{}

// NewPatternTracksParser creates and returns a PatternTracksParser.
func NewPatternTracksParser() *PatternTracksParser {
	return &PatternTracksParser{}
}

// Parse parses the given byte slice and returns a list of found
// tracks.
func (ptp *PatternTracksParser) Parse(b []byte) Tracks {
	tracks := Tracks{}
	pos := 50
	track := Track{}
	var err error
	for {
		if track, pos, err = ptp.parseTrack(b, pos); err != nil {
			break
		}
		tracks = append(tracks, track)
		if pos >= len(b) {
			break
		}
	}
	return tracks
}

func (ptp *PatternTracksParser) parseTrack(b []byte, pos int) (Track, int, error) {
	var id int
	var err error
	if id, pos, err = ptp.parseID(b, pos); err != nil {
		return Track{}, pos, err
	}

	var name string
	if name, pos, err = ptp.parseName(b, pos); err != nil {
		return Track{}, pos, err
	}

	return Track{
		ID:    id,
		Name:  name,
		Steps: ptp.parseSteps(b[pos : pos+16]),
	}, pos + 16, nil
}

func (ptp *PatternTracksParser) isValidIndex(b []byte, pos int) bool {
	return pos < len(b)
}

func (ptp *PatternTracksParser) parseID(b []byte, pos int) (int, int, error) {
	if !ptp.isValidIndex(b, pos+1) {
		return -1, pos, errors.New("unexpected input while parsing ID")
	}

	id, err := strconv.ParseInt(hex.EncodeToString(b[pos:pos+1]), 16, 0)
	return int(id), pos + 1, err
}

func (ptp *PatternTracksParser) parseName(b []byte, pos int) (string, int, error) {
	if !ptp.isValidIndex(b, pos+4) {
		return "", pos, errors.New("unexpected input while parsing length of Name")
	}

	var nameLength int64
	var err error
	if nameLength, err = strconv.ParseInt(hex.EncodeToString(b[pos:pos+4]), 16, 0); err != nil {
		return "", pos, err
	}

	pos = pos + 4
	if !ptp.isValidIndex(b, pos+int(nameLength)) {
		return "", pos, errors.New("unexpected input while parsing Name")
	}

	return string(b[pos : pos+int(nameLength)]), pos + int(nameLength), nil
}

func (ptp *PatternTracksParser) parseSteps(b []byte) Steps {
	steps := Steps{}
	for i := 0; i < 4; i++ {
		steps.Bars[i] = ptp.parseBar(b[4*i : 4*(i+1)])
	}

	return steps
}

func (ptp *PatternTracksParser) parseBar(b []byte) Bar {
	notes := [4]Note{}
	for i := 0; i < len(b); i++ {
		if note, _ := strconv.ParseInt(hex.EncodeToString(b[i:i+1]), 16, 0); note != 0 {
			notes[i] = true
		}
	}
	return Bar(notes)
}
