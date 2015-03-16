package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
)

const (
	fileIdentifierText = "SPLICE"
	widthHeaderField   = 10
	widthVersion       = 32
)

// FileIdentifier is what all valid Splice files should start with.
var FileIdentifier = append(
	[]byte(fileIdentifierText),
	bytes.Repeat([]byte{0}, widthHeaderField-len(fileIdentifierText))...,
)

// ErrHeaderMissing is returned if the input to decoder doesn't match
// what's expected.
var ErrHeaderMissing = fmt.Errorf(
	"expected file to start with identifier '%s'",
	string(FileIdentifier),
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("Error closing file: %v", err)
		}
	}()
	return Decode(f)
}

// Decode decodes a Pattern from a reader.
func Decode(r io.Reader) (p *Pattern, err error) {
	dataLen, err := DecodeHeader(r)
	if err != nil {
		err = fmt.Errorf("unable to decode identifier, %v", err)
		return
	}

	lr := io.LimitReader(r, int64(dataLen))

	p = NewPattern()
	if p.Version, p.Tempo, err = DecodeMeta(lr); err != nil {
		err = fmt.Errorf("unable to decode header, %v", err)
		return
	}
	for {
		var t *Track
		if t, err = DecodeTrack(r); err != nil {
			if err == io.EOF {
				err = nil
				break
			}
			return
		}
		p.Tracks = append(p.Tracks, t)
	}
	return
}

// DecodeHeader checks the file identifier is present and gets the length
// of the body in bytes.
func DecodeHeader(r io.Reader) (dataLen int32, err error) {
	// Check that the header is correct.
	p := make([]byte, widthHeaderField)
	if _, err = r.Read(p); err != nil {
		err = fmt.Errorf("unable to read identifier bytes, %v", err)
		return
	}
	if !bytes.Equal(p, FileIdentifier) {
		fmt.Printf("%s\n%s", p, FileIdentifier)
		err = ErrHeaderMissing
		return
	}

	// Read out the length of the data.
	// I'm not certain about the endianness as the others appear to be little,
	// but it made sense to me the body length was larger than a byte and 10
	// seems like a sensical size for the identifier.
	// It would actually be possible for it to be a big endian int64 given the
	// leading zeros, but I assumed everything was 32-bit if it's really from
	// the 90s, though it could even be lower than that.
	if err = binary.Read(r, binary.BigEndian, &dataLen); err != nil {
		err = fmt.Errorf("unable to read body length, %v", err)
	}
	return
}

// DecodeMeta decodes the start of the body to find the version and tempo.
func DecodeMeta(r io.Reader) (version string, tempo float32, err error) {
	// Read out the version.
	p := make([]byte, widthVersion)
	if _, err = r.Read(p); err != nil {
		err = fmt.Errorf("unable to read version bytes, %v", err)
		return
	}
	version = string(bytes.Replace(p, []byte{0}, []byte{}, -1))

	// Read out the tempo.
	if err = binary.Read(r, binary.LittleEndian, &tempo); err != nil {
		err = fmt.Errorf("unable to read tempo, %v", err)
	}
	return
}
