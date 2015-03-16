package drum

import (
	"bytes"
	"errors"
	"fmt"
)

// An Instrument is the representation of an instrument
// used by a track.
type Instrument struct {
	ID   uint8
	Name string
}

// Maximum length of an instrument name in bytes
// This is set to prevent excessive memory allocations on corrupt files.
const maxNameLen = 16

// ErrInstrumentName will be returned when the track decoder
// is unable to decode the instrument name.
var ErrInstrumentName = errors.New("unable to read the instument name")

// DecodeName will read the name from the Buffer stream.
func (i *Instrument) decodeName(r *bytes.Buffer, nameLen int) error {
	// Here we have a sanity check on the maximum name length.
	// Set in constant "maxNameLength"
	if nameLen > maxNameLen || nameLen < 0 {
		return ErrInstrumentName
	}

	// Read instrument name
	name := make([]byte, nameLen)

	// We have a typed bytes.Buffer, so it will always return the number of requested
	// bytes, see http://golang.org/pkg/bytes/#Buffer.Read
	n, err := r.Read(name)
	if err != nil {
		return err
	}
	if n != int(nameLen) {
		return ErrInstrumentName
	}
	i.Name = string(name)
	return nil
}

// String returns a human readable representation of the instrument.
func (i Instrument) String() string {
	return fmt.Sprintf("(%d) %s", i.ID, i.Name)
}
