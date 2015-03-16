// The decoder package provides high level management
// to decode splice files into Pattern(s) struct
package decoder

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/simcap/drum/splice"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	reader := splice.NewReader(content)
	header, tracks, err := reader.ReadAll()

	if err != nil {
		return nil, err
	}

	return NewPattern(header, tracks), nil
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version string
	Tempo   float32
	Tracks  []splice.Track
}

func NewPattern(h splice.Header, tracks []splice.Track) *Pattern {
	return &Pattern{h.Version(), h.Tempo, tracks}
}

// String returns a friendly printable output
// representing the Pattern
func (p Pattern) String() string {
	versioning := fmt.Sprintf("Saved with HW Version: %s", p.Version)

	tempo := strings.Replace(fmt.Sprintf("Tempo: %.1f", p.Tempo), ".0", "", 1)

	var alltracks bytes.Buffer
	for _, t := range p.Tracks {
		alltracks.WriteString(
			fmt.Sprintf("(%d) %s\t%s\n",
				t.Id,
				t.Name,
				t.Steps.Text("x", "-", "|"),
			),
		)
	}

	return fmt.Sprintf("%s\n%s\n%s", versioning, tempo, alltracks.String())
}
