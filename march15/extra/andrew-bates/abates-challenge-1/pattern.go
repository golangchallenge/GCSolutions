package drum

import (
	"fmt"
)

var (
	// ErrInvalidHeader indicates the header portion of the splice is not valid
	ErrInvalidHeader = fmt.Errorf("Invalid Header")

	// ErrInvalidVersion indicates the splice is encoded using a version that is
	// not supported by this parser
	ErrInvalidVersion = fmt.Errorf("Invalid Version")
)

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	version string
	tempo   float32
	tracks  []*Track
	err     error
}

// Decode will parse bytes from the Reader and popluate the Pattern
func (p *Pattern) Decode(r Reader) error {
	// TODO: verify packet/file structure
	//
	// It seems strange to me that the header is 13 bytes followed by a
	// single byte for the length.  This very much limits the length of a
	// splice file (max 255 bytes after the header/length).  It would
	// seem more reasonable to have a 10 byte header and 4 byte packet
	// length.  However, when using this configuration, the byte order
	// of the packet length field is wrong (little endian instead of
	// big endian).  It also seems unreasonable that a single field in the
	// file would have a different byte order than the rest of the file
	// Therefore, I'm puzzled a bit about the file structure
	//
	var header [13]byte
	var packetLen uint8

	r.Next(&header)
	// Verify header
	if r.Err() == nil && header != [13]byte{
		'S', 'P', 'L', 'I', 'C', 'E', 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00,
	} {
		return ErrInvalidHeader
	}

	r.Next(&packetLen)
	r.FixedString(32, &p.version)
	// Verify this is a version we can parse
	if r.Err() == nil && p.version != "0.708-alpha" &&
		p.version != "0.808-alpha" &&
		p.version != "0.909" {
		return ErrInvalidVersion
	}

	r.Next(&p.tempo)
	// The remaining bytes is the packet length minus the length of
	// the version string (32 bytes) and tempo (4 bytes)
	remainingBytes := int(packetLen) - 36
	for remainingBytes > 0 {
		track := &Track{}
		err := track.Decode(r)
		if err != nil {
			return err
		}

		// The length of a track is not fixed because the name is a variable length
		// field.  So we need to decrement remainingBytes by the actual length of
		// the track record
		remainingBytes -= track.length()
		p.tracks = append(p.tracks, track)
	}
	return r.Err()
}

// String returns the entire splice as a string using the following format:
//
//     Save with HW Version: <VERSION STRING>
//     Temp: <TEMPO>
//     <TRACK 0 String>
//     <TRACK 1 String>
//     ...
//     <TRACK N String>
//
// For example:
//    Saved with HW Version: 0.808-alpha
//    Tempo: 120
//    (0) kick     |x---|x---|x---|x---|
//    (1) snare    |----|x---|----|x---|
//    (2) clap     |----|x-x-|----|----|
//    (3) hh-open  |--x-|--x-|x-x-|--x-|
//    (4) hh-close |x---|x---|----|x--x|
//    (5) cowbell  |----|----|--x-|----|
func (p *Pattern) String() string {
	s := fmt.Sprintf("Saved with HW Version: %v\n", p.version)
	s += fmt.Sprintf("Tempo: %v\n", p.tempo)
	for _, track := range p.tracks {
		s += track.String()
	}
	return s
}
