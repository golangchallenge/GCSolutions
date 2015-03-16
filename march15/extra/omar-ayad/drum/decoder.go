package drum

import (
	"bytes"
	"encoding/binary"
	"io/ioutil"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
// TODO: implement

var versionBeginOffset int = 14
var tempoBeginOffset int = 46
var instrumentsBeginOffset int = 50

func DecodeFile(path string) (*Pattern, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	data := string(file)
	fileBegin := strings.Index(data, "SPLICE")
	fileEnd := strings.Index(data[1:], "SPLICE")
	if fileEnd == -1 {
		fileEnd = len(data)

	} else {
		fileEnd += 1
	}
	version := GetVersion(data, fileBegin, fileEnd, versionBeginOffset)
	tempo, _ := GetTempo(data, fileBegin, fileEnd, tempoBeginOffset)
	instruments := DecodeInstruments(data, fileBegin, fileEnd, instrumentsBeginOffset)
	p := &Pattern{version, tempo, instruments}
	return p, nil
}

// GetVersion(Binary File, File Begining, File End, Offset)
// gets the splice version of which the file
// was saved with
func GetVersion(binFile string, fileBegin int, fileEnd int, offset int) string {
	dataFromVersion := binFile[fileBegin+offset : fileEnd]
	versionEnd := strings.Index(dataFromVersion, "\x00")
	version := dataFromVersion[:versionEnd]
	return version
}

// GetTempo(Binary File, File Begining, File End, Offset)
// gets the tempo of which the splice file plays
func GetTempo(binFile string, fileBegin int, fileEnd int, offset int) (float32, error) {
	var tempo float32
	dataFromTempo := binFile[fileBegin+offset : fileEnd]
	tempoBytes := []byte(dataFromTempo[:4])
	buf := bytes.NewReader(tempoBytes)
	err := binary.Read(buf, binary.LittleEndian, &tempo)
	return tempo, err

}

// DecodeInstrument(Binary File, Upper Limit, Offset)
// decodes a single instrument into
// an instrument struct and returns it along with
// the index of the next instrument
func DecodeInstrument(binString string, limit int, offset int) (Instrument, int) {
	idOfInstrument := int64(binString[0])
	dataFromInstrumentName := binString[5:]
	var nameOfInstrument string = ""
	var i int = 0
	characters := dataFromInstrumentName[i]
	for characters != 1 && characters != 0 {
		nameOfInstrument += string(characters)
		i++
		characters = dataFromInstrumentName[i]
	}
	var beats [16]bool
	for j := 0; j < 16; j++ {
		if (i + j + offset) >= limit {
			break
		}
		if int64(dataFromInstrumentName[i+j]) == 1 {
			beats[j] = true
		} else if int64(dataFromInstrumentName[i+j]) == 0 {
			beats[j] = false
		}

	}
	instrument := Instrument{idOfInstrument, nameOfInstrument, beats}
	return instrument, i + 21
}

// DecodeInstruments(binary file, File Begining, File End, Offset)
// decodes all of the instruments and returns an array of structs
// containing all instruments of the file
func DecodeInstruments(binFile string, fileBegin int, fileEnd int, offset int) []Instrument {
	dataFromFirstInstrument := binFile[fileBegin+offset : fileEnd]
	var instruments []Instrument
	var i int = 0
	var indexOffset int = 0
	var instrumentTemp Instrument
	for i+fileBegin+offset < fileEnd {
		instrumentTemp, indexOffset = DecodeInstrument(dataFromFirstInstrument[i:], fileEnd, fileBegin+offset+i)
		instruments = append(instruments, instrumentTemp)
		i += indexOffset
	}
	return instruments
}

// the format method, formats the pattern struct
// into a printable manner
func (p *Pattern) Format() (string, error) {
	x := *p
	var formatedString string = "Saved with HW Version: " + x.version + "\n"

	if int(x.tempo)*10 == int(x.tempo*10) {
		formatedString += "Tempo: " + strconv.FormatInt(int64(x.tempo), 10) + "\n"
	} else {
		formatedString += "Tempo: " + strconv.FormatFloat(float64(x.tempo), 'f', 1, 32) + "\n"
	}

	for i := 0; i < len(x.instruments); i++ {
		formatedString += "(" + strconv.FormatInt(x.instruments[i].id, 10) + ") " + x.instruments[i].name + "\t"
		for j := 0; j < 16; j++ {
			if j%4 == 0 {
				formatedString += "|"
			}
			if x.instruments[i].beats[j] {
				formatedString += "x"
			} else {
				formatedString += "-"
			}
		}
		formatedString += "|\n"
	}
	return formatedString, nil

}

// AddCowbell method adds x extra
// steps to the existing cowbell instrument
func (p *Pattern) AddCowbell(x int) error {
	for i := 0; i < len(p.instruments); i++ {
		if p.instruments[i].name == "cowbell" {
			cowbellBeats := p.instruments[i].beats
			var empty []int
			for j := 0; j < 16; j++ {
				if cowbellBeats[j] == false {
					empty = append(empty, j)
				}
			}
			if len(empty) <= x {
				p.instruments[i].beats = [16]bool{true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true}
			} else {
				r := rand.New(rand.NewSource(time.Now().UnixNano()))
				var randIndex int
				for j := 0; j < x; j++ {
					randIndex = r.Int() % len(empty)
					p.instruments[i].beats[empty[randIndex]] = true
					empty = empty[:randIndex+copy(empty[randIndex:], empty[randIndex+1:])]
				}
			}
			break
		}
	}
	return nil
}

// A representation of a single instrument
type Instrument struct {
	id    int64
	name  string
	beats [16]bool
}

// A representation of the whole splice file
type Pattern struct {
	version     string
	tempo       float32
	instruments []Instrument
}
