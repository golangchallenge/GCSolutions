package drum

import (
	"io"
	"os"
)

// Decoder reads and decodes drum machine object fron an input stream.
type Decoder struct {
	r io.Reader // Input stream of splice object.
}

// NewDecoder returns a new decoder that reads from r.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

// Decode reads the binary-encoded value from its input d.r and return the parsed
// Pattern.
func (d *Decoder) Decode() (*Pattern, error) {
	// Header of "SPLICE", with padding, plus pattern size.
	h := make([]byte, lenHeader+1)
	if _, err := d.r.Read(h); err != nil {
		return nil, err
	}

	// Bytes containing Pattern.
	b := make([]byte, h[posPatternSize])
	if _, err := d.r.Read(b); err != nil {
		return nil, err
	}

	p := &Pattern{size: h[posPatternSize]}
	return p, p.UnmarshalBinary(b)
}

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	r, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return NewDecoder(r).Decode()
}
