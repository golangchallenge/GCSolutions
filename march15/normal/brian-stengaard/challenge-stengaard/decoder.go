package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

var (
	magic = [...]byte{'S', 'P', 'L', 'I', 'C', 'E', 0, 0, 0, 0, 0, 0, 0}
)

const (
	magicAddr = 0x0
	magicLen  = 0xd

	dataLenAddr = magicAddr + magicLen
	dataLenLen  = 0x1

	versionAddr = dataLenAddr + dataLenLen
	versionLen  = 0x20

	tempoAddr = versionAddr + versionLen
	tempoLen  = 0x4

	trackAddr = tempoAddr + tempoLen
)

// DecodeReader reads a Pattern from r, parses it and returns the Pattern.
// An error is returned if the a Pattern cannot be parsed from r.
func DecodeReader(r io.Reader) (*Pattern, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	ph := &struct {
		Magic   [magicLen]byte
		DataLen uint8
		Name    [nameLenLen]byte
		Tempo   float32
	}{}

	err = binary.Read(bytes.NewReader(data), binary.LittleEndian, ph)
	if err != nil {
		return nil, err
	}

	if ph.Magic != magic {
		return nil, errors.New("bad header")
	}
	// slice down the data so we have only what we are going to use.
	if int(ph.DataLen) < len(data)-(magicLen+dataLenLen) {
		data = data[:magicLen+dataLenLen+ph.DataLen]
	}

	// len(magic) + 1 to include the length field itself
	if len(data)-(magicLen+dataLenLen) != int(ph.DataLen) {
		return nil, fmt.Errorf("bad length - read %d expected %d", ph.DataLen, len(data)-len(magic))
	}
	p := &Pattern{
		Version: readCString(data[versionAddr:trackAddr]),
	}
	binary.Read(bytes.NewReader(data[tempoAddr:]), binary.LittleEndian, &p.Tempo)

	chunk := data[trackAddr:]
	for len(chunk) > 0 {
		n, track, err := readTrack(chunk)
		if err != nil {
			return nil, err
		}
		p.Tracks = append(p.Tracks, track)
		if len(chunk) < n {
			break
		}
		chunk = chunk[n:]
	}

	return p, nil

}

const (
	trackIDAddr = 0x0
	trackIDLen  = 0x4

	nameLenAddr = trackIDAddr + trackIDLen
	nameLenLen  = 0x1

	trackNameAddr = nameLenAddr + nameLenLen

	stepsLen = 0x10
)

// Read a track from the front of b. Returns how many bytes of b
// was consumed and a track. Returns an error if the track record in b
// is not consistent.
func readTrack(b []byte) (n int, t *Track, err error) {
	if len(b) < trackNameAddr+stepsLen {
		return 0, t, errors.New("buffer to short to contain track")
	}

	t = &Track{}
	th := &struct {
		ID      uint32
		NameLen uint8
	}{}
	err = binary.Read(bytes.NewReader(b), binary.LittleEndian, th)
	if err != nil {
		return 0, nil, err
	}
	t.ID = th.ID
	l := int(th.NameLen)
	if l > len(b) {
		return 0, nil, fmt.Errorf("corrupt track record. Bad Length %d > %d)", th.NameLen, len(b))
	}
	if th.NameLen == 0 {
		return 0, nil, fmt.Errorf("track record has zero length name")
	}

	t.Name = string(b[trackNameAddr : trackNameAddr+l])
	// re-slice in original buffer. Skips allocation and no
	// reference to the original buffer leaks.
	t.Steps = b[trackNameAddr+l : trackNameAddr+l+stepsLen]
	return trackIDLen + nameLenLen + l + stepsLen, t, nil
}

// Reads a C-Style (\0 terminated) string from the beginning of b.
func readCString(b []byte) string {
	for i := 0; i < len(b); i++ {
		if b[i] == 0 {
			return string(b[:i])
		}
	}
	return string(b)
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version string
	Tempo   float32

	Tracks []*Track
}

// String generates a string version of p
func (p *Pattern) String() string {
	r := []string{
		fmt.Sprintf("Saved with HW Version: %s", p.Version),
		fmt.Sprintf("Tempo: %v", p.Tempo),
	}

	for i := range p.Tracks {
		r = append(r, p.Tracks[i].String())
	}
	return strings.Join(r, "\n") + "\n"
}

// Track is a sound played whenever the
// Steps is not zero.
type Track struct {
	ID    uint32
	Name  string
	Steps []byte
}

func (t Track) String() string {
	out := []byte{}
	add := func(b byte) {
		out = append(out, b)
	}
	for i := 0; i < len(t.Steps); i++ {
		if i%4 == 0 {
			add('|')
		}

		if t.Steps[i] == 0 {
			add('-')
		} else {
			add('x')
		}
	}
	add('|') // finish up
	return fmt.Sprintf("(%d) %s\t%s", t.ID, t.Name, string(out))
}

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err

	}
	defer f.Close()

	return DecodeReader(f)
}
