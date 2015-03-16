package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {

	// Create a struct to hold the pattern
	p := &Pattern{}
	// open the binary file
	file, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	defer file.Close()

	fInfo, err := os.Stat(path)
	// read some binary data
	buffer := make([]byte, fInfo.Size())
	_, err = file.Read(buffer)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	// try to just read the magic string
	r := bytes.NewReader(buffer)
	MagicString := make([]byte, 13)
	_, err = io.ReadAtLeast(r, MagicString, 13)
	if err != nil {
		return nil, err
	}

	// try to read the version
	Version := make([]byte, 33)
	_, err = io.ReadAtLeast(r, Version, 33)
	if err != nil {
		return nil, err
	}

	// try to read the tempo
	tmp := make([]byte, 4)
	_, err = io.ReadAtLeast(r, tmp, 4)
	if err != nil {
		return nil, err
	}
	var tempo Tempo
	tempoReader := bytes.NewReader(tmp)
	binary.Read(tempoReader, binary.LittleEndian, &tempo)

	var id TrackId
	tracks := make([]Track, 0)
	done := false
	for !done {
		// try to read the first track id
		_, err = io.ReadAtLeast(r, tmp, 4)
		if err != nil {
			done = true
			//return nil, err
		}
		idReader := bytes.NewReader(tmp)
		binary.Read(idReader, binary.LittleEndian, &id)

		// try to read the first track name
		// first we need to get the length (this should be at byte 54)
		lenReader := make([]byte, 1)
		_, err = io.ReadAtLeast(r, lenReader, 1)
		if err != nil {
			done = true
			//return nil, err
		}
		var nameLength uint8
		nameLengthReader := bytes.NewReader(lenReader)
		binary.Read(nameLengthReader, binary.LittleEndian, &nameLength)

		name := make([]byte, nameLength)
		_, err = io.ReadAtLeast(r, name, int(nameLength))
		if err != nil {
			done = true
			//return nil, err
		}
		// try to read the first track data
		data := make([]byte, 16)
		_, err = io.ReadAtLeast(r, data, 16)
		if err != nil {
			done = true
			break
		}
		var t Track
		t.Id = id
		t.Name = name
		t.Data = data
		tracks = append(tracks, t)
	}

	p.MagicString = MagicString
	p.Version = bytes.Trim(Version, "\x00")
	p.Version = bytes.TrimLeft(p.Version, "\xc5")
	p.Version = bytes.TrimLeft(p.Version, "W")
	p.Tempo = tempo
	p.Tracks = tracks
	return p, nil
}

// A MagicString represents the SPLICE format magic string.  The magic string is
// encoded in the first 13 bytes of the file.
type MagicString []byte

// A Version represents the SPLICE software version used to write this
// file
type Version []byte

// A Tempo represents a littleEndian encoded tempo (beats per minute)
type Tempo float32

// A TrackId is assigned to each Track
type TrackId uint32

// A TrackName is assigned to each Track
type TrackName []byte

// A TrackData is the encoded 0/1 for each of the 16 steps in a track
type TrackData []byte

type Track struct {
	Id   TrackId
	Name TrackName
	Data TrackData
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
// TODO: implement
type Pattern struct {
	MagicString MagicString
	Version     Version
	Tempo       Tempo
	Tracks      []Track
}

func (p *Pattern) String() string {
	out := fmt.Sprintf("Saved with HW Version: %s\n", p.Version)
	out = out + fmt.Sprintf("Tempo: %g\n", p.Tempo)
	for i := range p.Tracks {
		out = out + fmt.Sprintf("(%d) %s\t", p.Tracks[i].Id, p.Tracks[i].Name)
		for j := range p.Tracks[i].Data {
			if j%4 == 0 {
				out = out + fmt.Sprintf("|")
			}
			if p.Tracks[i].Data[j] == 0 {
				out = out + fmt.Sprintf("-")
			} else if p.Tracks[i].Data[j] == 1 {
				out = out + fmt.Sprintf("x")
			}
		}
		out = out + fmt.Sprintf("|\n")
	}
	return out
}
