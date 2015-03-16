package drum

import (
	"io/ioutil"
)

// A PatternReader reads a file and takes the appropriate actions
// to begin parsing.
type PatternReader struct {
	filePath string
}

// NewPatternReader creates and returns a PatternReader given
// a file path to a pattern file.
func NewPatternReader(filePath string) *PatternReader {
	return &PatternReader{
		filePath: filePath,
	}
}

// Read pulls in a file and starts the parsing process, returning
// a complete Pattern if successful.
func (pr *PatternReader) Read() (*Pattern, error) {
	b, err := ioutil.ReadFile(pr.filePath)
	if err != nil {
		return nil, err
	}

	parser := NewPatternParser()

	return parser.Parse(b)
}
