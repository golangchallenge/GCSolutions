package drum

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
// TODO: implement
func DecodeFile(path string) (*Pattern, error) {
	return Open(path)
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
// TODO: implement
type Pattern struct {
	Version string
	Tempo   float32
	Tracks  []*Track
}

type Track struct {
	ID    int32
	Name  string
	Beats []byte
}
