package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

var (
	header           = "SPLICE"
	headerLenght     = 14
	versionLenght    = 32
	stepLenght       = 16
	quartersPerTrack = 4
	stepPerQuarter   = 4
	versionLabel     = "Saved with HW Version: "
	tempoLabel       = "Tempo: "
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	var readedBytes int

	p := NewPattern()

	file, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	p.Filename = filepath.Base(path)

	n, err := readHeader(file, &p.Lenght)

	if err != nil {
		return p, err
	}

	n, err = readVersion(file, &p.Version)

	if err != nil {
		return p, err
	}

	readedBytes += n

	n, err = readTempo(file, &p.Tempo)

	if err != nil {
		return p, err
	}

	readedBytes += n

	n, err = readTracks(file, &p.Tracks, int(p.Lenght), readedBytes)

	if err != nil {
		return p, err
	}

	return p, nil
}

// Step is an internal renaming of the bool value
type Step bool

// Quarter is fourth of a Tack
type Quarter struct {
	Steps []Step
}

// Track is the struct that represent a track of a Pattern
type Track struct {
	Number   uint8
	ID       uint8
	Name     string
	Quarters []Quarter
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Filename string
	Lenght   int64
	Version  string
	Tempo    float32
	Tracks   []Track
}

// NewPattern creates a new Pattern
func NewPattern() *Pattern {
	p := &Pattern{}
	p.Tracks = []Track{}
	return p
}

// NewTrack creates a new Track
func NewTrack() *Track {
	t := &Track{}
	t.Quarters = []Quarter{}
	return t
}

// NewQuarter creates a new Quarter
func NewQuarter(steps []byte) (*Quarter, error) {
	var err error
	q := &Quarter{}

	q.Steps = make([]Step, len(steps), len(steps))

	for k, s := range steps {

		b := s == byte(1)
		q.Steps[k] = Step(b)
	}

	return q, err
}

// Consumes the SPLICE file format header of 14 bytes and check if the the consumed header is SPLICE with the message
// lenght at the end f the header encoded in binary.BigEndian
func readHeader(r io.Reader, length *int64) (int, error) {
	buf := make([]byte, len(header), len(header))

	n, err := r.Read(buf)

	if n != len(header) || err != nil {
		return n, fmt.Errorf("Error reading header. The header bust be \"%v\" with a lenght of %v. Readed %v bytes. Error was: %v\n", header, len(header), n, err)
	}

	header := string(buf)

	if header != header {
		return n, fmt.Errorf("The reader header %v is not SPLICE", header)
	}

	err = binary.Read(r, binary.BigEndian, length)

	return headerLenght, err
}

func readVersion(r io.Reader, version *string) (int, error) {
	bf := make([]byte, versionLenght, versionLenght)

	n, err := r.Read(bf)

	if n != versionLenght || err != nil {
		return n, fmt.Errorf("Error reading version. The version bust have a lenght of %v. Readed %v bytes. Error was: %v\n", versionLenght, n, err)
	}

	*version = string(bytes.TrimRight(bf, "\x00"))

	return n, nil
}

func readTempo(r io.Reader, tempo *float32) (int, error) {
	err := binary.Read(r, binary.LittleEndian, tempo)

	return 4, err
}

// Read and decode the track from the reader r. It will read the tracks until the fileLenght is not reached. If after
// reaching the fileLenght the are more bytes to read, those bytes will be ommited
// Returns the readed bytes, and an error if any
func readTracks(r io.Reader, tracks *[]Track, fileLenght int, readedBytes int) (int, error) {
	var track *Track
	var tBytes int // Number of readed bytes
	var nTrack int

	tBytes = readedBytes

	for {

		// Stop reading tracks when the specified fileLenght is reached
		if tBytes >= fileLenght {
			break
		}
		track = NewTrack()
		n, err := readTrack(r, track, fileLenght, readedBytes)

		if err != nil {
			return n, fmt.Errorf("Error reading track #%v. Nested error: %v", nTrack, err)
		}
		*tracks = append(*tracks, *track)

		// Increment reader bytes
		tBytes += n
		//Increment number of trackss
		nTrack++
	}

	return int(tBytes), nil
}

func readTrack(r io.Reader, track *Track, fileLenght int, readedBytes int) (int, error) {
	var tnbr int // Total number of bytes readed so far
	var trackID byte
	var trackNameLength int32

	err := binary.Read(r, binary.LittleEndian, &trackID)
	if err != nil {
		return -1, fmt.Errorf("Error reading trackId. %v", err)
	}
	tnbr++
	track.ID = uint8(trackID)

	err = binary.Read(r, binary.BigEndian, &trackNameLength)

	if err != nil {
		return -1, fmt.Errorf("Error reading trackNameLength. %v", err)
	}
	tnbr = tnbr + 4

	trackName := make([]byte, trackNameLength, trackNameLength)
	n, err := r.Read(trackName)

	if n != int(trackNameLength) || err != nil {
		return -1, fmt.Errorf("Error reading trackname with length of %v. Only readed %v. %v", trackNameLength, n, err)
	}
	track.Name = string(trackName)
	tnbr = tnbr + n

	steps := make([]byte, stepLenght, stepLenght)

	n, err = r.Read(steps)

	if n != stepLenght || err != nil {
		return -1, fmt.Errorf("Error reading steps with length of %v. Only readed %v. %v", stepLenght, n, err)
	}

	tnbr = tnbr + n

	track.Quarters = make([]Quarter, quartersPerTrack, quartersPerTrack)
	for i := 0; i < quartersPerTrack; i++ {
		q, err := NewQuarter(steps[i*stepPerQuarter : (i*stepPerQuarter)+stepPerQuarter])

		if err != nil {
			break
		}

		track.Quarters[i] = *q
	}

	return tnbr, err
}

func (q Quarter) String() string {
	var buf bytes.Buffer
	buf.WriteString("|")

	for _, s := range q.Steps {
		if s {
			buf.WriteString("x")
		} else {
			buf.WriteString("-")
		}
	}

	return buf.String()
}

func (t Track) String() string {
	var buf bytes.Buffer

	buf.WriteString("(")
	buf.WriteString(strconv.Itoa(int(t.ID)))
	buf.WriteString(") ")
	buf.WriteString(t.Name)
	buf.WriteString("\t")

	for _, q := range t.Quarters {
		buf.WriteString(q.String())
	}

	buf.WriteString("|")

	return buf.String()
}

func (p Pattern) String() string {
	var buf bytes.Buffer

	buf.WriteString(versionLabel)
	buf.WriteString(p.Version)
	buf.WriteString("\n")
	buf.WriteString(tempoLabel)
	buf.WriteString(fmt.Sprintf("%g", p.Tempo))
	buf.WriteString("\n")
	for _, t := range p.Tracks {
		buf.WriteString(t.String())
		buf.WriteString("\n")
	}

	return buf.String()
}
