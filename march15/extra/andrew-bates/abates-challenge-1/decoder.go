package drum

import (
	"io"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	p := &Pattern{}
	r := newFileReader(path)
	if closer, ok := r.(io.Closer); ok {
		defer closer.Close()
	}
	return p, p.Decode(r)
}

// DecodeByteArray takes a splice defined as a byte array and returns a pointer
// to an initialized Pattern.  If any parse errors are encountered while
// reading the byte array, than an error is returned
func DecodeByteArray(splice []byte) (*Pattern, error) {
	p := &Pattern{}
	r := newByteArrayReader(splice)
	return p, p.Decode(r)
}
