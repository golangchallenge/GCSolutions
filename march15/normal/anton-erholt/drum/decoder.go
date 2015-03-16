package drum

import (
	"encoding/binary"
	"errors"
	"os"
)

const (
	MIN_TRACK_SIZE    = 21 // bytes
	TRACKS_OFFSET     = 36 // bytes
	UPPER_TEMPO_LIMIT = 1000.0
)

// A binaryTrack is a fairly low level reperesentation of a track.
type binaryTrack struct {
	id        int32
	strlen    byte
	name      []byte
	steps     [16]byte
	byte_size uint32
}

// Checks if bt is considered a valid track.
func (bt *binaryTrack) isValid() (bool, error) {
	if uint32(bt.strlen) <= 0 {
		return false, errors.New("Too short name on track.")
	}

	if bt.id < 0 {
		return false, errors.New("Found a negative id on track.")
	}

	if bt.byte_size <= MIN_TRACK_SIZE {
		return false, errors.New("Invalid byte size on track.")
	}
	return true, nil
}

// A fairly low level representation of a Pattern.
// Used as an intermediate representation for easier separation of
// transformation steps.
type binaryPattern struct {
	splice          [6]byte // The the text SPLICE in ascii/UTF-8
	flags           uint32  // big endian, should be zero.
	remaining_bytes uint32  // big endian
	version_string  [32]byte
	tempo           float32 // little endian
	tracks          []binaryTrack
}

// Checks if bp is considered a valid pattern.
func (bp *binaryPattern) isValid() (bool, error) {

	// Should start with 'SPLICE'
	switch {
	case bp.splice[0] != 0x53,
		bp.splice[1] != 0x50,
		bp.splice[2] != 0x4c,
		bp.splice[3] != 0x49,
		bp.splice[4] != 0x43,
		bp.splice[5] != 0x45:
		return false, errors.New("Invalid file header. Couldn't parse SPLICE.")
	}

	// Four '\0' bytes
	if bp.flags != 0 {
		return false, errors.New("Invalid file header. Flags shouldn't be set.")
	}

	if bp.tempo <= 0 || bp.tempo >= UPPER_TEMPO_LIMIT {
		return false, errors.New("Corrupted data. Unable to parse a valid tempo.")
	}

	// Make sure remaining bytes is correct.
	// 32 (version) + 4 (tempo) + sum(track.bytesize for track in bp.tracks)
	var current_size uint32 = TRACKS_OFFSET
	for _, track := range bp.tracks {
		current_size += track.byte_size
		if ok, err := track.isValid(); !ok {
			return false, err
		}
	}

	if current_size != bp.remaining_bytes {
		return false, errors.New("Invalid file header. Wrong number of bytes remaining.")
	}
	return true, nil
}

// Converts a binaryPattern to a Pattern.
func (bp *binaryPattern) toPattern() (p *Pattern) {
	p = new(Pattern)

	// Convert headers
	n := 0
	for bp.version_string[n] != byte(0x0) {
		n++
	}
	p.version = string(bp.version_string[:n])
	p.tempo = bp.tempo

	// Convert tracks
	p.tracks = make(map[int]Track, len(bp.tracks))
	p.printOrder = make([]int, 0)
	for _, bt := range bp.tracks {
		var t Track
		t.name = string(bt.name[:int(bt.strlen)])
		for i := 0; i < len(t.steps); i++ {
			t.steps[i] = bt.steps[i] == 0x1
		}

		var id int = int(bt.id)
		p.tracks[id] = t
		p.printOrder = append(p.printOrder, id)
	}
	return p
}

// A binaryReader is a type for easier reading of binary data into structs.
// It wraps the 'error'-idiom described on the golang blog.
// See http://blog.golang.org/errors-are-values
type binaryReader struct {
	file *os.File
	err  error
}

// Constructs a new BinaryReader from an *os.File.
func newBinaryReader(file *os.File) (bpr *binaryReader) {
	bpr = new(binaryReader)
	bpr.file = file
	bpr.err = nil
	return bpr
}

// Reads the data byte by byte into data.
// Analogous with http://golang.org/pkg/encoding/binary/#Read
func (bpr *binaryReader) Read(order binary.ByteOrder, data interface{}) {
	if bpr.err == nil {
		bpr.err = binary.Read(bpr.file, order, data)
		return
	}
}

// Returns any error reported so far by the underlying binary.Read().
func (bpr *binaryReader) Error() error {
	return bpr.err
}

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern, which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return nil, err
	}

	bpr := newBinaryReader(file)

	// Read binary pattern
	var bp binaryPattern
	bpr.Read(binary.BigEndian, &bp.splice)
	bpr.Read(binary.LittleEndian, &bp.flags)
	bpr.Read(binary.BigEndian, &bp.remaining_bytes)
	bpr.Read(binary.BigEndian, &bp.version_string)
	bpr.Read(binary.LittleEndian, &bp.tempo)
	var bytes_left = bp.remaining_bytes - TRACKS_OFFSET

	bp.tracks = make([]binaryTrack, 0)

	// Read tracks
	for bytes_left > 0 {
		var bt binaryTrack
		bpr.Read(binary.LittleEndian, &bt.id)
		bpr.Read(binary.BigEndian, &bt.strlen)
		bt.name = make([]byte, int(bt.strlen), int(bt.strlen))
		bpr.Read(binary.BigEndian, &bt.name)
		bpr.Read(binary.BigEndian, &bt.steps)
		bt.byte_size = MIN_TRACK_SIZE + uint32(bt.strlen)

		bp.tracks = append(bp.tracks, bt)

		bytes_left -= bt.byte_size
	}

	// Error handling
	if bpr.err != nil {
		return nil, err
	}

	if ok, err := bp.isValid(); !ok {
		return nil, err
	}

	p := bp.toPattern()
	return p, nil
}
