// The splice package contains the splice file reader
// along with details on the format of splice files
package splice

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"unicode"
)

const (
	headerSize       int64 = 50
	measureSize            = 4
	maxTrackNameSize       = 20
	MagicNumber            = "SPLICE"
)

type Header struct {
	MagicNumber [6]byte
	_           [8]byte
	Vers        [12]byte
	_           [20]byte
	Tempo       float32
}

// Version return the string of the splice version
// with only printable characters
func (h Header) Version() string {
	var buff bytes.Buffer
	for _, b := range h.Vers {
		if unicode.IsPrint(rune(b)) {
			buff.WriteByte(b)
		}
	}
	return buff.String()
}

type Track struct {
	Id    uint8
	Name  []byte
	Steps Steps
}

// Steps represents the 16 slots for sound on the track
type Steps [16]byte

// String returns the default display of the steps content
func (s Steps) String() string {
	return s.Text("1", "0", "")
}

// Text allows to print the steps in a friendly manner
// using an up string to represent a sound, down for a silence
// and sep as the measure separator
func (s Steps) Text(up, down, sep string) string {
	var result = bytes.NewBufferString(sep)
	var section bytes.Buffer

	for i, b := range s {
		if b == 0 {
			section.WriteString(down)
		} else {
			section.WriteString(up)
		}
		if (i+1)%measureSize == 0 {
			section.WriteString(sep)
			result.Write(section.Bytes())
			section.Reset()
		}
	}
	return result.String()
}

type Reader struct {
	*bytes.Reader
	contentSize int
}

var (
	ErrMissingMagicNumber = fmt.Errorf("expecting magic number '%s' in content's header", MagicNumber)
)

func NewReader(content []byte) *Reader {
	return &Reader{bytes.NewReader(content), len(content)}
}

func (r *Reader) ReadAll() (Header, []Track, error) {
	h, err := r.GetHeader()
	if err != nil {
		return h, []Track{}, err
	}

	ts, err := r.GetTracks()
	if err != nil {
		return h, ts, err
	}

	return h, ts, nil
}

func (r *Reader) GetHeader() (Header, error) {
	r.Seek(0, 0)
	header := Header{}

	if !r.hasMagicNumberNext() {
		return header, ErrMissingMagicNumber
	}

	r.read(&header, binary.LittleEndian)

	return header, nil
}

func (r *Reader) GetTracks() ([]Track, error) {
	r.Seek(headerSize, 0)
	tracks := []Track{}

	for r.Len() > 0 && !r.hasMagicNumberNext() {
		t, err := r.getTrack()
		if err != nil {
			return tracks, err
		}
		tracks = append(tracks, t)
	}

	return tracks, nil
}

func (r *Reader) getTrack() (Track, error) {
	var id uint8
	r.read(&id)

	var nameSize uint32
	r.read(&nameSize)

	if nameSize > maxTrackNameSize {
		return Track{}, fmt.Errorf(
			"found track name too large (limited to %d) around byte %d in content",
			maxTrackNameSize,
			r.position(),
		)
	}

	name := make([]byte, nameSize)
	r.read(&name)

	var steps Steps
	r.read(&steps)

	return Track{id, name, steps}, nil
}

func (r *Reader) read(data interface{}, order ...binary.ByteOrder) {
	var o binary.ByteOrder
	if len(order) > 0 {
		o = order[0]
	} else {
		o = binary.BigEndian
	}

	err := binary.Read(r, o, data)
	if err != nil {
		fmt.Fprintf(os.Stdout, "error trying to read a %T around byte %d in content: %v\n", data, r.position(), err)
		os.Exit(1)
	}
}

func (r *Reader) hasMagicNumberNext() bool {
	defer r.Seek(-int64(r.Len()), 2)

	next := make([]byte, 6)
	r.read(next)

	return bytes.Equal(next, []byte(MagicNumber))
}

func (r *Reader) position() int {
	return r.contentSize - r.Len()
}
