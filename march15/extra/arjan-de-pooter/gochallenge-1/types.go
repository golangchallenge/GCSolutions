package drum

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version string
	Tempo   float32
	Tracks  []*Track
}

// Track is a representation of a single track in a pattern
type Track struct {
	ID    int
	Name  string
	Steps [16]bool
}
