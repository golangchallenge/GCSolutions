package drum

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

// magic word at the beginning of splice file
var headerMagic []byte = []byte("SPLICE")

// how many steps a track contains
const trackSteps = uint8(16)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (p *Pattern, err error) {
	p = &Pattern{
		trackMap: make(map[uint8]int),
	}

	// open file
	var f *os.File
	f, err = os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	br := bufio.NewReader(f)

	// read header and check it
	if err = checkHeader(br); err != nil {
		return
	}

	var readBytes int
	var totalRead uint64

	// read expected length of the pattern
	var patternLen uint64
	if err = binary.Read(br, binary.BigEndian, &patternLen); err != nil {
		return
	}

	// pattern version
	version := make([]byte, 32)
	if readBytes, err = br.Read(version); err != nil {
		return
	}
	totalRead += uint64(readBytes)
	p.Version = strings.Trim(string(version), " \x00")

	// tempo
	if err = binary.Read(br, binary.LittleEndian, &p.Tempo); err != nil {
		return
	}
	totalRead += 4 // 4 bytes read

	// tracks
	for err == nil {
		var track Track
		track, readBytes, err = readTrack(br)
		if err == nil {
			err = p.AddTrack(track)

			totalRead += uint64(readBytes)
		}
	}

	if err == io.EOF {
		err = nil
	}

	if err == nil && totalRead != patternLen {
		err = fmt.Errorf("wrong number of bytes read, expected '%d', got '%d'",
			patternLen,
			totalRead,
		)
	}

	return
}

// checkHeader reads header value of splice file and checks if it matches
// expected magic value
func checkHeader(r io.Reader) error {
	hlen := len(headerMagic)
	header := make([]byte, hlen, hlen)

	if _, err := r.Read(header); err != nil {
		return err
	}
	if !bytes.Equal(header, headerMagic) {
		return fmt.Errorf("splice header is corrupted, expected '%s', got '%s'",
			headerMagic,
			header,
		)
	}

	return nil
}

// readTrack tries to read one track from current position of reader
func readTrack(r io.Reader) (t Track, totalRead int, err error) {
	// read track id
	if err = binary.Read(r, binary.BigEndian, &t.ID); err != nil {
		return
	}
	totalRead++

	// read track name length
	var nameLen uint32
	if err = binary.Read(r, binary.BigEndian, &nameLen); err != nil {
		return
	}
	totalRead += 4

	var read int
	// read track name
	name := make([]byte, nameLen)
	if read, err = r.Read(name); err != nil {
		return
	}
	totalRead += read
	t.Name = string(name)

	// read steps
	buf := make([]byte, 1)
	for i := uint8(0); i < trackSteps; i++ {
		if read, err = r.Read(buf); err != nil {
			return
		}
		totalRead += read
		if buf[0] == 1 {
			t.Steps = t.Steps + 1<<i

		}
	}

	return
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version  string
	Tempo    float32
	tracks   []Track
	trackMap map[uint8]int

	m sync.Mutex
}

// GetTrack returns pointer to track by id
func (p *Pattern) GetTrack(id uint8) (t *Track, ok bool) {
	var i int
	i, ok = p.trackMap[id]
	if ok {
		t = &p.tracks[i]
	}

	return
}

// GetTracks returns tracks of the pattern
func (p *Pattern) GetTracks() []Track {
	return p.tracks
}

// AddTrack adds a track to pattern. Returns error if there is already
// existing track with the same id
func (p *Pattern) AddTrack(t Track) error {
	p.m.Lock()

	if _, ok := p.trackMap[t.ID]; ok {
		return fmt.Errorf("track with id '%d' already exists in pattern", t.ID)
	}

	p.tracks = append(p.tracks, t)
	p.trackMap[t.ID] = len(p.tracks) - 1
	p.m.Unlock()

	return nil
}

func (p *Pattern) String() string {
	return DefaultPatternEncoder.Encode(p)
}

// Track is the high level representation of the drum track
type Track struct {
	ID    uint8
	Name  string
	Steps uint16
}

// SetStep sets or unsets a step for the track.
// Panics if index is out of range
func (t *Track) SetStep(index uint8, on bool) {
	if index > trackSteps {
		panic(fmt.Sprintf("step index must be less than %d, got '%d'",
			trackSteps,
			index,
		))
	}

	if on {
		t.Steps = t.Steps | (1 << index)

	} else {
		t.Steps = t.Steps & ^(1 << index)

	}
}

func (t *Track) String() string {
	return DefaultTrackEncoder.Encode(t)
}
