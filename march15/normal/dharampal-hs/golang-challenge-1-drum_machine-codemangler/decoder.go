package drum

import (
	"bufio"
	"fmt"
	"os"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	p := &Pattern{}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := file.Close(); err != nil {
			panic(err)
		}
	}()

	bufferedReader := bufio.NewReader(file)

	header, errHeader := parseHeader(bufferedReader)
	if errHeader != nil {
		return nil, errHeader
	}
	p.header = header

	tracks, errTracks := parseTrackCollection(bufferedReader, p.header.contentSize())
	if errTracks != nil {
		return nil, errTracks
	}
	p.tracks = tracks

	return p, nil
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	header *Header
	tracks []Track
}

func (pattern Pattern) String() string {
	patternString := fmt.Sprintf("%s\n", pattern.header)
	for _, track := range pattern.tracks {
		patternString += fmt.Sprintf("%s\n", track)
	}
	return patternString
}
