package drum

// A PatternParser is responsible for parsing the individual pieces
// of a pattern file.
type PatternParser struct {
	headerParser *PatternHeaderParser
	tracksParser *PatternTracksParser
}

// NewPatternParser returns an initialized PatternParser.
func NewPatternParser() *PatternParser {
	return &PatternParser{
		headerParser: NewPatternHeaderParser(),
		tracksParser: NewPatternTracksParser(),
	}
}

// Parse calls its child parsers and returns, if successful,
// a parsed pattern.
func (pp *PatternParser) Parse(b []byte) (*Pattern, error) {
	header, err := pp.headerParser.Parse(b)
	if err != nil {
		return nil, err
	}

	return &Pattern{
		version: header.GetVersion(),
		tempo:   header.GetTempo(),
		tracks:  pp.tracksParser.Parse(b),
	}, nil
}
