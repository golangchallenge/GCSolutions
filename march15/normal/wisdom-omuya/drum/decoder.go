package drum

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// DecodeFile decodes the drum machine file found at the provided path and
// returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	r, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening %v: %v", path, err)
	}
	defer r.Close()

	h, err := decodeHeader(r)
	if err != nil {
		return nil, fmt.Errorf("header decode error: %v", err)
	}

	// holds the number of bytes encoded afterwards
	n := int32(hdrSize - 10 /* company name */ - 4 /* encoded data size */)

	// holds the tracks in the pattern
	t := []*Track{}

	for {
		dt, err := decodeTrack(r)
		if err != nil {
			return nil, fmt.Errorf("track decode error: %v", err)
		}

		t = append(t, dt)

		n += int32(dt.Size)

		if n == h.size {
			// if we've exhausted all valid encoded bytes, we're done
			break
		}
	}

	p := &Pattern{
		Header: h,
		Tracks: t,
	}

	return p, nil
}

// decodeHeader decodes the first 50 bytes of the encoded pattern binary
// and returns the pattern's header and any error encountered.
//
// A PatternHeader's layout can be seen as follows:
//
// Byte Range	| Information
// =============|=======================================================
//	0 - 9		| machine's company name (null terminated, includes padding)
//	10 - 13		| size of valid encoded data (data starts at byte 14)
//	14 - 45		| machine's hardware version (null terminated, includes padding)
//	46 - 49		| IEEE-754 32 bit floating point number holding tempo
//
//  The first 14 bytes hold the PatternHeader - containing both the company
//  name and the size of the entire encoding.
//
func decodeHeader(r io.Reader) (*PatternHeader, error) {
	c, err := readString(r)
	if err != nil {
		return nil, fmt.Errorf("error decoding company: %v", err)
	}

	p := computePadding(10 /* bytes 0 - 9 for name */, c)

	// company name can not be more than 10 bytes
	if p < 0 {
		return nil, fmt.Errorf("company name too long - %vB (max 10B)", len(c))
	}

	s, err := decodeSize(r, p)
	if err != nil {
		return nil, fmt.Errorf("error decoding size: %v", err)
	}

	v, err := readString(r)
	if err != nil {
		return nil, fmt.Errorf("error decoding version: %v", err)
	}

	p = computePadding(32 /* bytes 14 - 45 for drum version */, v)

	// version string can not be more than 32 bytes
	if p < 0 {
		return nil, fmt.Errorf("drum version too long - %vB (max 32B)", len(v))
	}

	t, err := decodeTempo(r, p)
	if err != nil {
		return nil, fmt.Errorf("error decoding tempo: %v", err)
	}

	h := &PatternHeader{
		company: c,
		size:    s,
		Version: v,
		Tempo:   t,
	}

	return h, nil
}

// decodeTempo decodes the Pattern size given reader, r, and padding, p.
func decodeSize(r io.Reader, p int) (int32, error) {
	// advance reader to end of version string padding
	_, err := advanceReader(r, p)
	if err != nil {
		return 0, fmt.Errorf("advance error: %v", err)
	}

	// now read the size of valid encoded message
	return readInt32(r, bigEndian)
}

// decodeTempo decodes the Pattern tempo given reader, r, and padding, p.
func decodeTempo(r io.Reader, p int) (float32, error) {
	// advance reader to end of version string padding
	_, err := advanceReader(r, p)
	if err != nil {
		return 0, fmt.Errorf("advance error: %v", err)
	}

	var t float32

	// TODO(wao) probably more efficient ways of doing this w/o re-implementing
	// IEEE-754? for now just use the binary package
	return t, binary.Read(r, binary.LittleEndian, &t)
}

// decodeTrack decodes a track from r and returns it and any error encountered.
//
// A Track's layout can be seen as follows:
//
// Byte Range	| Information
// =============|================================
//	0			| track ID
//	1 - 4		| length of the track name
//	5 - ?		| holds the track name
//	? - ? + 15	| holds the steps of the track
//
//  The first 5 bytes hold the TrackHeader - containing both the track id and
//  the length of the track name.
//
//	? is the offset at which the track name ends.
//
func decodeTrack(r io.Reader) (*Track, error) {
	h, err := decodeTrackHeader(r)
	if err != nil {
		return nil, fmt.Errorf("error decoding header: %v", err)
	}

	n, err := readBytes(r, h.Size)
	if err != nil {
		return nil, fmt.Errorf("error decoding name: %v", err)
	}

	// read steps of the track
	s, err := readBytes(r, 16)
	if err != nil {
		return nil, fmt.Errorf("error decoding steps: %v", err)
	}

	// size of entire track struct
	l := int32(5 /* track header */ + len(n) /* track name */ + 16 /* track steps */)

	t := &Track{
		Header: h,
		Name:   string(n),
		Steps:  s,
		Size:   l,
	}

	return t, nil
}

// decodeTrackHeader decodes a TrackHeader from r.
func decodeTrackHeader(r io.Reader) (*TrackHeader, error) {
	// read one byte representing track id
	id, err := readBytes(r, 1)
	if err != nil {
		return nil, fmt.Errorf("id error: %v", err)
	}

	// read the length of the track name
	s, err := readInt32(r, bigEndian)
	if err != nil {
		return nil, fmt.Errorf("name length error: %v", err)
	}

	// sanity check for track name
	if s < 1 {
		return nil, fmt.Errorf("name length is %vB", s)
	}

	h := &TrackHeader{
		Id:   id[0],
		Size: s,
	}

	return h, nil
}
