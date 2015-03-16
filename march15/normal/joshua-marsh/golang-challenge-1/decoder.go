package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
)

var (
	// ErrNotSpliceFile returned during DecodeFile if the file doesn't
	// contain the header information necessary to parse the file.
	ErrNotSpliceFile = errors.New("not a splice file")
)

// Pattern is the high level representation of the drum pattern
// contained in a .splice file.
type Pattern struct {
	Version string
	Tempo   float32
	Tracks  []*Track
}

// String implements the fmt.Stringer interface.
func (p *Pattern) String() string {
	buf := &bytes.Buffer{}
	buf.WriteString("Saved with HW Version: ")
	buf.WriteString(p.Version)
	buf.WriteString("\nTempo: ")
	buf.WriteString(fmt.Sprintf("%v", p.Tempo))
	for _, track := range p.Tracks {
		buf.WriteByte('\n')
		buf.WriteString(track.String())
	}
	buf.WriteByte('\n')
	return buf.String()
}

// Track is a high level representation of a track for a drum pattern.
type Track struct {
	ID    int32
	Name  string
	Steps []bool
}

// String implements the fmt.Stringer interface.
func (t *Track) String() string {
	buf := &bytes.Buffer{}
	for i, step := range t.Steps {
		if i%4 == 0 { // Write out the bars (every 4th).
			buf.WriteString("|")
		}
		if step {
			buf.WriteString("x")
		} else {
			buf.WriteString("-")
		}
	}
	buf.WriteString("|")
	return fmt.Sprintf("(%v) %v\t%v", t.ID, t.Name, buf.String())
}

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point
// to the rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return decodePattern(file)
}

// decodePattern creates a Pattern from the data in the given reader.
func decodePattern(r io.Reader) (*Pattern, error) {
	// Verify that this is a splice file.
	buf := make([]byte, 32)
	if _, err := io.ReadFull(r, buf[:13]); err != nil {
		return nil, ErrNotSpliceFile
	}
	if bytes.Compare(buf[:6], []byte("SPLICE")) != 0 {
		return nil, ErrNotSpliceFile
	}

	// Get the remaining number of bytes of the rest of the contents.
	var size uint8
	if err := binary.Read(r, binary.LittleEndian, &size); err != nil {
		return nil, ErrNotSpliceFile
	}

	// Get the version of this file.
	p := &Pattern{}
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, ErrNotSpliceFile
	}
	p.Version = string(buf[:bytes.IndexByte(buf, 0x00)])

	// Get the tempo.
	if err := binary.Read(r, binary.LittleEndian, &p.Tempo); err != nil {
		return nil, ErrNotSpliceFile
	}

	// Read the rest of the tracks, making sure to only read the number
	// of bytes specified in the header.
	remaining := int(size) - 36
	for remaining > 0 {
		track, err := decodeTrack(r)
		if err != nil {
			return nil, err
		}
		p.Tracks = append(p.Tracks, track)
		// We know how much was read based on the length of the name of
		// the track.
		remaining -= 21 + len(track.Name)
	}
	return p, nil
}

// decodeTrack creates a Track from the data in the given reader.
func decodeTrack(r io.Reader) (*Track, error) {
	t := &Track{
		Steps: make([]bool, 16),
	}
	// Get the ID.
	if err := binary.Read(r, binary.LittleEndian, &t.ID); err != nil {
		return nil, err
	}

	// Get the length of the name and then the name.
	var l uint8
	if err := binary.Read(r, binary.LittleEndian, &l); err != nil {
		return nil, err
	}
	buf := make([]byte, l)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}
	t.Name = string(buf)

	// Get each of the steps.
	for x := 0; x < 16; x++ {
		if _, err := io.ReadFull(r, buf[:1]); err != nil {
			return nil, err
		}
		if buf[0] == 0x01 {
			t.Steps[x] = true
		}
	}
	return t, nil
}

// EncodeFile writes the given drum machine to the given file.
func EncodeFile(p *Pattern, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return encodePattern(p, file)
}

// errWriter is a helper to simplify error handling for writes when
// encoding.
type errWriter struct {
	w   io.Writer
	err error
}

// Write implements the io.Writer interface.
func (w *errWriter) Write(b []byte) (int, error) {
	if w.err != nil {
		return 0, w.err
	}
	var n int
	n, w.err = w.w.Write(b)
	return n, w.err
}

// encodePattern writes the binary equivalent of the given patther to
// the given io.Writer.
func encodePattern(p *Pattern, w io.Writer) error {
	ew := &errWriter{w: w}
	// Write the header.
	ew.Write([]byte("SPLICE\000\000\000\000\000\000\000"))

	// Write the size of the rest of the content.
	var size uint8 = 36
	for _, t := range p.Tracks {
		size += 21 + uint8(len(t.Name))
	}
	binary.Write(ew, binary.LittleEndian, &size)

	// Write the version.
	ew.Write([]byte(p.Version))
	// Pad with 0's up to 32 bytes.
	for x := 0; x < 32-len(p.Version); x++ {
		ew.Write([]byte{0x00})
	}

	// Write the tempo.
	binary.Write(ew, binary.LittleEndian, &p.Tempo)

	// Write the tracks.
	for _, t := range p.Tracks {
		binary.Write(ew, binary.LittleEndian, t.ID)
		binary.Write(ew, binary.LittleEndian, uint8(len(t.Name)))
		ew.Write([]byte(t.Name))
		for _, step := range t.Steps {
			if step {
				ew.Write([]byte{0x01})
			} else {
				ew.Write([]byte{0x00})
			}
		}
	}

	return ew.err
}
