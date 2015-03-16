package drum

import (
	"bytes"
	"encoding/binary"
	"io"
	"os"
	"strconv"
)

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	version string
	tempo   float32
	tracks  []Track
}

// Track contains the id, name and steps of an instrument.
// A step is either 0 (do nothing) or 1 (play sound).
type Track struct {
	id    uint32
	name  string
	steps [16]uint8
}

const spliceHeader = "SPLICE"

// decoder is the stuct containing the decoding state.
type decoder struct {
	r      io.Reader // file being read
	tmp    [256]byte // buffer for reading file bits
	pat    *Pattern  // pattern being parsed
	trk    *Track    // track being parsed
	remain int       // number of unread bytes in SPLICE file body
}

// A FormatError reports that the input is not a valid SPLICE file.
type FormatError string

func (e FormatError) Error() string { return "splice: invalid format: " + string(e) }

// checkHeader checks if the file starts with "SPLICE".
func (d *decoder) checkHeader() error {
	_, err := io.ReadFull(d.r, d.tmp[:len(spliceHeader)])
	if err != nil {
		return err
	}
	if string(d.tmp[:len(spliceHeader)]) != spliceHeader {
		return FormatError("missing SPLICE header")
	}
	return nil
}

// parseLength extracts the body length.
// It assumes it is stored as a unsigned byte.
func (d *decoder) parseLength() error {
	_, err := io.ReadFull(d.r, d.tmp[:8])
	if err != nil {
		return err
	}
	n := uint8(d.tmp[7])
	d.remain = int(n)
	return nil
}

// parseVersion extracts the nul-terminated version string.
func (d *decoder) parseVersion() error {
	_, err := io.ReadFull(d.r, d.tmp[:32])
	if err != nil {
		return err
	}
	d.tmp[31] = 0
	n := bytes.Index(d.tmp[:32], []byte{0})
	d.pat.version = string(d.tmp[:n])
	d.remain -= 32
	return nil
}

// parseTempo extracts the tempo (32 bits float, little endian).
func (d *decoder) parseTempo() error {
	_, err := io.ReadFull(d.r, d.tmp[:4])
	if err != nil {
		return err
	}
	buf := bytes.NewReader(d.tmp[:4])
	if err := binary.Read(buf, binary.LittleEndian, &d.pat.tempo); err != nil {
		return FormatError("cannot decode tempo")
	}
	d.remain -= 4
	return nil
}

// parseTrackD extracts a track id. This implementation assumes an ID is
// a 32 bits little endian unsigned integer, but it could be an uint8 as well,
// as the examples only show the first byte being used.
func (d *decoder) parseTrackID() error {
	_, err := io.ReadFull(d.r, d.tmp[:4])
	if err != nil {
		return err
	}
	d.trk.id = binary.LittleEndian.Uint32(d.tmp[:4])
	d.remain -= 4
	return nil
}

// parseTrackName extracts a track name, which is a length-prefixed byte string.
// Length is assumed to be stored in an unsigned byte.
func (d *decoder) parseTrackName() error {
	_, err := io.ReadFull(d.r, d.tmp[:1])
	if err != nil {
		return err
	}
	d.remain--
	n := uint8(d.tmp[0])
	if int(n) > d.remain {
		return FormatError("track name too long")
	}
	_, err = io.ReadFull(d.r, d.tmp[:n])
	if err != nil {
		return err
	}
	d.trk.name = string(d.tmp[:n])
	d.remain -= int(n)
	return nil
}

// parseTrackSteps extracts the 16 steps, which are 0x00 or 0x01 bytes.
// No conversion done, the byte array is just copied into the track.
func (d *decoder) parseTrackSteps() error {
	_, err := io.ReadFull(d.r, d.trk.steps[:])
	if err != nil {
		return err
	}

	d.remain -= 16
	return nil
}

// parseTrack extracts a track (id, name and steps).
func (d *decoder) parseTrack() error {
	if err := d.parseTrackID(); err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return err
	}
	if err := d.parseTrackName(); err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return err
	}
	if err := d.parseTrackSteps(); err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return err
	}
	return nil
}

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
// Implemention inspired from some parts of image/png decoding
func DecodeFile(path string) (*Pattern, error) {
	p := &Pattern{}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	d := &decoder{r: file, pat: p}
	if err := d.checkHeader(); err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return nil, err
	}
	if err := d.parseLength(); err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return nil, err
	}
	if err := d.parseVersion(); err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return nil, err
	}
	if err := d.parseTempo(); err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return nil, err
	}

	for d.remain > 0 {
		d.trk = &Track{}
		if err := d.parseTrack(); err != nil {
			if err == io.EOF {
				err = io.ErrUnexpectedEOF
			}
			return nil, err
		}
		d.pat.tracks = append(d.pat.tracks, *d.trk)
	}
	return p, nil
}

// Displays a pattern in human-readable form.
// Got rid of fmt.Sprintf and string concatenations.
func (p *Pattern) String() string {
	var buf bytes.Buffer

	buf.WriteString("Saved with HW Version: ")
	buf.WriteString(p.version)
	buf.WriteString("\nTempo: ")
	buf.WriteString(strconv.FormatFloat(float64(p.tempo), 'f', -1, 32))
	buf.WriteString("\n")

	for _, t := range p.tracks {
		buf.WriteString("(")
		buf.WriteString(strconv.FormatUint(uint64(t.id), 10))
		buf.WriteString(") ")
		buf.WriteString(t.name)
		buf.WriteString("\t|")

		for i, step := range t.steps {
			if step == 0 {
				buf.WriteString("-")
			} else {
				buf.WriteString("x")
			}
			// print bar every 4 steps
			if (i+1)%4 == 0 {
				buf.WriteString("|")
			}
		}
		buf.WriteString("\n")
	}
	return buf.String()
}
