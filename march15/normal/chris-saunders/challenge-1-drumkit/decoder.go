package drum

import (
	"bufio"
	"bytes"
	"io"
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
	buff := bufio.NewReader(file)
	p.Header = ExtractHeader(buff)
	for true {
		if _, err := buff.Peek(1); err == io.EOF {
			break
		}
		sample, err := ExtractSample(buff)
		if err != nil {
			break
		}
		p.Tracks = append(p.Tracks, sample)
	}

	return p, nil
}

// Pattern is the struct that holds onto
// header information and tracks loaded
// from a splice file
type Pattern struct {
	Header Header
	Tracks []Sample
}

func (p Pattern) String() string {
	buff := bytes.NewBufferString(p.Header.String())
	buff.WriteString("\n")
	for _, t := range p.Tracks {
		buff.WriteString(t.String())
		buff.WriteString("\n")
	}
	return string(buff.Bytes())
}
