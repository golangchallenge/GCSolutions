package drum

import "fmt"

// Pattern contains a Header with some info and Tracks with steps and instrument.
type Pattern struct {
	Header *Header
	Tracks []*Track
}

// NewPattern returns a new Pattern with an initialized Header
func NewPattern() *Pattern {
	p := &Pattern{
		Header: &Header{},
	}

	return p
}

// AddTrack appends a Track to the Tracks slice.
func (p *Pattern) AddTrack(t *Track) {
	p.Tracks = append(p.Tracks, t)
}

func (p *Pattern) String() string {
	s := ""

	if p.Header != nil {
		s = fmt.Sprintf("%s%s\n", s, p.Header.String())
	}

	for _, t := range p.Tracks {
		s = fmt.Sprintf("%s%s\n", s, t.String())
	}

	return s
}
