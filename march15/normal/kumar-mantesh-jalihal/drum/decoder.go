package drum

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math"
	"strconv"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	p := &Pattern{}

	d, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println("unable to read file at: ", path)
		return nil, err
	}

	verEnc := d[14:25]
	tempoEnc := d[44:50]
	tracksEnc := d[50:]

	p.ver = formatVersion(verEnc)
	p.tempo, err = formatTempo(tempoEnc)
	p.tracks = formatTracks(tracksEnc)
	if err != nil {
		fmt.Println("Unable to process tempo:", err)
		return nil, err
	}

	return p, nil
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	ver    string
	tempo  float32
	tracks []track
}

type track struct {
	id    int
	name  string
	steps []bool
}

// Returns the formated version number of the pattern
func formatVersion(d []byte) string {
	v := ""
	for i := 0; i < len(d); i++ {
		if d[i] != 0 {
			v += fmt.Sprintf("%c", d[i])
		}
	}
	return string(v)
}

// Returns the formated tempo of the pattern
func formatTempo(d []byte) (float32, error) {
	var tmpo float32
	tmpoEnc := ""
	for i := len(d) - 1; i >= 0; i-- {
		tmpoEnc += fmt.Sprintf("%x", d[i])
		if d[i] == 0 {
			tmpoEnc += "0"
		}
	}
	tmpoEncHex, err := hex.DecodeString(tmpoEnc)
	if err != nil {
		fmt.Println("Hex Decode failed:", err)
		return 0, err
	}
	buf := bytes.NewReader(tmpoEncHex)
	err = binary.Read(buf, binary.BigEndian, &tmpo)
	if err != nil {
		fmt.Println("binary.Read failed:", err)
		return 0, err
	}
	return tmpo, nil
}

// Returns the formated track arrays of the pattern
func formatTracks(d []byte) []track {
	// There are four parts in d, each requires to be decoded one at a time.
	const (
		id      = iota
		nameLen = iota
		name    = iota
		step    = iota
	)

	var turn int
	var trackNameLen int
	var stepCount int
	var trackName string
	var tracks []track
	var t track

	for i := 0; i < len(d); i++ {
		switch turn {
		case id:
			t = track{}
			t.steps = make([]bool, 16, 16)
			t.id, _ = strconv.Atoi(fmt.Sprintf("%d", d[i]))
			turn = nameLen
			break
		case nameLen:
			if d[i] == 0 {
				continue
			} else {
				trackNameLen, _ = strconv.Atoi(fmt.Sprintf("%d", d[i]))
				turn = name
			}
			break
		case name:
			trackName += fmt.Sprintf("%c", d[i])
			trackNameLen--
			if trackNameLen == 0 {
				t.name = trackName
				trackName = ""
				turn = step
			}
			break
		case step:
			stepVal := false
			if d[i] == 0 {
				stepVal = false
			} else {
				stepVal = true
			}
			t.steps[stepCount] = stepVal
			stepCount++
			if stepCount == 16 {
				stepCount = 0
				tracks = append(tracks, t)
				turn = id
			}
			break
		}
	}
	return tracks
}

// Implements Stringer interface to return the formatted pattern.
func (p *Pattern) String() string {
	s := ""
	s += "Saved with HW Version: " + p.ver + "\n"
	s += "Tempo: " + fmt.Sprintf("%v", p.tempo) + "\n"

	for _, t := range p.tracks {
		s += "(" + fmt.Sprintf("%d", t.id) + ") "
		s += t.name + "\t"
		for k, n := range t.steps {
			if math.Mod(float64(k), 4) == 0 {
				s += "|"
			}
			if n == true {
				s += "x"
			} else {
				s += "-"
			}
		}
		s += "|\n"
	}
	return s
}
