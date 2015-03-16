package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

// Header is the SPLICE magic number to indicate that the file is a splice file.
const Header string = "SPLICE"
const (
	// The byte location of the content size
	ContentSizeLoc = 13
	// How big is each step in the system
	StepSize = 16
	// Size of the Header block in bytes.
	headerBlockSize = 14
	// Size of the Version block in bytes
	VersionBlockSize = 32
)

// This would be used for encoding.
var header []byte = []byte{'S', 'P', 'L', 'I', 'C', 'E', 0, 0, 0, 0, 0, 0, 0}

func ztString(buf *bytes.Buffer) (string, int) {

	if buf.Len() == 0 {
		return "", 0
	}
	s, err := buf.ReadString(byte(0)) // Only return err if it reaches EOF before finding 0
	n := len(s)
	if err != nil {
		return s, n
	}
	s = s[:len(s)-1]
	return s, n
}

func (t *Track) String() string {
	pat := make([]rune, 0, StepSize+(StepSize/4))
	for i, v := range t.Steps {
		if i%4 == 0 {
			pat = append(pat, '|')
		}
		if v {
			pat = append(pat, 'x')
		} else {
			pat = append(pat, '-')
		}
	}
	pat = append(pat, '|')
	return fmt.Sprintf("(%d) %s\t%s", t.Id, t.Name, string(pat))
}
func (t *Track) Decode(r *bytes.Buffer) error {
	var id int32
	err := binary.Read(r, binary.LittleEndian, &id)
	if err != nil {
		return err
	}

	size, err := r.ReadByte()
	if err != nil {
		return err
	}
	name := string(r.Next(int(size)))
	pat := make([]int8, StepSize, StepSize)
	binary.Read(r, binary.LittleEndian, &pat)
	for i, v := range pat {
		t.Steps[i] = v == 1
	}
	t.Id = int(id)
	t.Name = name

	return nil
}
func (t Tracks) String() string {
	strs := make([]string, 0, len(t))
	for _, v := range t {
		strs = append(strs, v.String())
	}
	return strings.Join(strs, "\n")
}
func decodeTracks(b *bytes.Buffer) (Tracks, error) {
	trks := make(Tracks, 0, 2)
	for b.Len() != 0 {
		t := &Track{}
		err := t.Decode(b)
		if err != nil {
			return trks, err
		}
		trks = append(trks, t)
	}
	return trks, nil
}

func (p *Pattern) String() string {
	return fmt.Sprintf("Saved with HW Version: %s\nTempo: %g\n%s\n", p.Version, p.Tempo, p.Tracks)
}
func validateHeader(b []byte) (contentSize int, valid bool) {
	if len(b) < headerBlockSize {
		return 0, false
	}
	contentSize = int(b[ContentSizeLoc])
	buff := bytes.NewBuffer(b)
	headr, _ := ztString(buff)
	if Header != headr {
		return 0, false
	}
	return contentSize, true
}

func (p *Pattern) decode(b *bytes.Buffer) error {

	version, n := ztString(b)
	if n < VersionBlockSize {
		b.Next(VersionBlockSize - n) // eat up extra bytes
	}

	var tempo float32
	err := binary.Read(b, binary.LittleEndian, &tempo)
	if err != nil {
		return err
	}
	trks, err := decodeTracks(b)
	if err != nil {
		return err
	}
	p.Tempo = tempo
	p.Version = version
	p.Tracks = trks
	return nil
}

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	p := &Pattern{}
	f, err := os.Open(path)
	if err != nil {
		return p, err
	}
	var b []byte = make([]byte, headerBlockSize)

	_, err = f.Read(b)
	if err != nil {
		f.Close()
		return p, err
	}
	var contentSize int
	var ok bool
	if contentSize, ok = validateHeader(b); !ok {
		f.Close()
		return p, fmt.Errorf("Not a SPLICE File (%s)", err)
	}
	b = make([]byte, contentSize)
	_, err = f.Read(b)
	if err != nil {
		f.Close()
		return p, err
	}
	f.Close()

	buff := bytes.NewBuffer(b)
	err = p.decode(buff)

	return p, err
}

// Tracks is a representation of the tracks contained in a Pattern
type Track struct {
	// Steps are when the beat needs to be played.
	Steps [StepSize]bool
	// The name of the Instrument Track
	Name string
	// The Id of the Track
	Id int
}

type Tracks []*Track

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	// Version of the Hardware the this pattern was saved on
	Version string
	// Tempo is the Beats per minute that the pattern should be played at
	Tempo float32
	// Tracks of the Pattern for each instrument
	Tracks Tracks
}
