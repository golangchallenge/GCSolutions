// Package drum is supposed to implement the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
package drum

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type header struct {
	ChunkID         [6]byte
	Padding1        [7]byte
	Unknown1        [1]byte
	HardwareVersion [31]byte
	Unknown2        [2]byte
	TempoDecimal    byte
	Tempo           byte
	Unknown3        byte
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Tempo           int
	TempoDecimal    int
	HardwareVersion string
	Tracks
}

// NewPattern returns an empty pattern.
func NewPattern() *Pattern {
	p := new(Pattern)
	p.Tracks = make([]Track, 0)
	return p
}

func (p Pattern) String() string {
	bpm := fmt.Sprint(p.Tempo)
	if p.TempoDecimal != 0 {
		bpm = fmt.Sprintf("%v.%v", p.Tempo, p.TempoDecimal)
	}
	s := fmt.Sprintf("Saved with HW Version: %s\nTempo: %v\n%v", p.HardwareVersion, bpm, p.Tracks)
	return s
}

func (p Pattern) header() header {
	h := header{}
	h.ChunkID = [6]byte{'S', 'P', 'L', 'I', 'C', 'E'}
	for i, r := range p.HardwareVersion {
		h.HardwareVersion[i] = byte(r)
	}
	if p.HardwareVersion == "0.808-alpha" {
		if p.TempoDecimal != 0 {
			h.TempoDecimal = byte(p.TempoDecimal + 200)
		}
		h.Tempo = byte(p.Tempo * 2)
	}
	return h
}

// NewPatternFromBackup creates a pattern structure by parsing
// a backup file's human-readible text data.
func NewPatternFromBackup(s string) (*Pattern, error) {
	scanner := bufio.NewScanner(strings.NewReader(s))
	p := NewPattern()
	for lineNum := 0; scanner.Scan(); lineNum++ {
		line := scanner.Text()
		switch lineNum {
		case 0:
			p.HardwareVersion = parseHardwareVersion(line)
		case 1:
			var err error
			p.Tempo, p.TempoDecimal, err = parseTempo(line)
			if err != nil {
				return p, err
			}
		default:
			t, err := parseTrack(line)
			if err != nil {
				return p, err
			}
			p.Tracks = append(p.Tracks, t)
		}
	}
	return p, nil
}

func parseHardwareVersion(line string) string {
	s := strings.TrimLeft(line, "Saved with HW Version: ")
	return s
}

func parseTempo(line string) (tempo, tempoDecimal int, err error) {
	s := strings.TrimLeft(line, "Tempo: ")
	match := tempoRe.FindStringSubmatch(s)
	tempo, err = strconv.Atoi(match[1])
	if err != nil {
		return 0, 0, err
	}
	tempoDecimal, err = strconv.Atoi(match[2])
	if err != nil {
		return tempo, 0, nil
	}
	return tempo, tempoDecimal, nil
}

var idRe = regexp.MustCompile(`\((\d+)\) `)
var tempoRe = regexp.MustCompile("(\\d+).?(\\d+)?")
var beatRe = regexp.MustCompile(`([x-]{4})\|`)

func parseTrack(line string) (Track, error) {
	id, line, err := parseTrackID(line)
	if err != nil {
		return Track{}, err
	}
	name, line := parseTrackName(line)
	bars, line, err := parseBar(line, 4)
	if err != nil {
		return Track{}, err
	}
	return Track{Name: name, ID: id, Sequence: bars}, nil
}

func parseTrackID(line string) (id byte, leftTrimmedLine string, err error) {
	idMatch := idRe.FindStringSubmatch(line)
	if len(idMatch) != 2 {
		return 0, "", fmt.Errorf("No track ID parsed from line: '%v'", line)
	}
	n, err := strconv.Atoi(idMatch[1])
	if err != nil {
		return id, leftTrimmedLine, err
	}
	leftTrimmedLine = strings.TrimLeft(line, idMatch[0])
	return byte(n), leftTrimmedLine, nil
}

var nameRe = regexp.MustCompile(`([\w-]+)\s+\|`)

func parseTrackName(line string) (name, leftTrimmedLine string) {
	s := strings.SplitN(line, "|", 2)
	name = strings.TrimRight(s[0], " \t")
	return name, s[1]
}

func parseBar(line string, numMeasures int) (bar []byte, leftTrimmedLine string, err error) {
	measureMatches := beatRe.FindAllStringSubmatch(line, numMeasures)
	for i := 0; i < numMeasures; i++ {
		measure := measureMatches[i][1]
		beats, err := parseBeats(measure)
		if err != nil {
			return beats, leftTrimmedLine, err
		}
		bar = append(bar, beats...)
		line = strings.TrimLeft(line, measureMatches[i][0])
	}
	return bar, line, nil
}

func parseBeats(measure string) ([]byte, error) {
	var beats []byte
	for _, beat := range measure {
		switch beat {
		case onBeat:
			beats = append(beats, 1)
		case offBeat:
			beats = append(beats, 0)
		default:
			return beats, fmt.Errorf("Unknown beat character '%v' encountered parsing measure '%v'", beat, measure)
		}
	}
	return beats, nil
}

// A Track represents a named, identified drum sequence.
type Track struct {
	ID       byte
	Name     string
	Sequence []byte
}

// NewTrack returns an empty, initialized track.
func NewTrack() *Track {
	t := new(Track)
	t.Sequence = make([]byte, 16)
	return t
}

func (t Track) encode() []byte {
	b := []byte{t.ID, 0, 0, 0}
	b = append(b, byte(len(t.Name)))
	b = append(b, []byte(t.Name)...)
	b = append(b, t.Sequence...)
	return b
}

const (
	separator rune = '|'
	onBeat    rune = 'x'
	offBeat   rune = '-'
	errorRune rune = '?'
)

func (t Track) String() string {
	var b bytes.Buffer
	b.WriteString(fmt.Sprintf("(%d) %s\t", t.ID, t.Name))
	for i := 0; i < len(t.Sequence); i++ {
		if i%4 == 0 {
			b.WriteString(string(separator))
		}
		switch t.Sequence[i] {
		case 1:
			b.WriteString(string(onBeat))
		case 0:
			b.WriteString(string(offBeat))
		default:
			b.WriteString(string(errorRune))
		}
	}
	b.WriteString(string(separator))
	return b.String()
}
