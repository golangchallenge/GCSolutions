package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strings"
)

const magicNumber = "SPLICE" //identifier file
const maxSteps = 16          //steps by track
const measureSteps = 4       //used to measure steps of track

// Track it's a sound track
type Track struct {
	ID    int
	Name  string
	Steps [16]bool
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version string
	Tempo   float32
	Tracks  []Track
}

// patterRawHeader is the low level representation of the file header
type patternRawHeader struct {
	Magic    [13]byte
	FileSize byte
	Version  [32]byte
	Tempo    float32
}

// Track by ID
func (p *Pattern) Track(id int) *Track {
	for _, track := range p.Tracks {
		if track.ID == id {
			return &track
		}
	}
	return nil
}

// AddTrack to pattern
func (p *Pattern) AddTrack(track Track) {
	p.Tracks = append(p.Tracks, track)
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (p *Pattern) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer
	var bufTracks bytes.Buffer

	for _, track := range p.Tracks {

		binary.Write(&bufTracks, binary.LittleEndian, int32(track.ID))
		bufTracks.WriteByte(byte(len(track.Name)))
		//limit the header to 256 bytes
		io.CopyN(&bufTracks, strings.NewReader(track.Name), 256)

		for _, step := range track.Steps {
			if step {
				bufTracks.WriteByte(byte(1))
			} else {
				bufTracks.WriteByte(byte(0))
			}
		}
	}

	header := patternRawHeader{
		Tempo: p.Tempo,
	}
	header.FileSize = byte(bufTracks.Len() + len(header.Version) /*version*/ + 4 /*int32*/)

	copy(header.Magic[:], []byte(magicNumber))
	copy(header.Version[:], []byte(p.Version))

	if err := binary.Write(&buf, binary.LittleEndian, header); err != nil {
		return nil, err
	}

	if _, err := bufTracks.WriteTo(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// UnmarshalBinary implements the encoding.BinaryMarshaler interface.
func (p *Pattern) UnmarshalBinary(data []byte) error {
	var tracks []Track
	var header patternRawHeader

	buf := bytes.NewReader(data)
	if buf.Len() == 0 {
		return errors.New("drum.Pattern.UrmashalBinary: no data")
	}

	binary.Read(buf, binary.LittleEndian, &header)
	if string(header.Magic[0:len(magicNumber)]) != magicNumber {
		return errors.New("drum.Pattern.UnmarshalBinary: invalid magic id")
	}

	header.FileSize -= byte(len(header.Version) + measureSteps /*float32*/)
	if len(data) < int(header.FileSize)+len(header.Magic)+1 /*file size byte*/ {
		return errors.New("drum.Pattern.UnmarshalBinary: invalid file length")
	}

	for {
		var trackIDBuf uint32
		startTrackSize := buf.Len()
		if err := binary.Read(buf, binary.LittleEndian, &trackIDBuf); err != nil {
			return err
		}

		sizeName, err := buf.ReadByte()
		if err != nil {
			return err
		}
		trackNameBuf := make([]byte, sizeName)
		if _, err := buf.Read(trackNameBuf); err != nil {
			return err
		}

		steps := make([]byte, maxSteps)
		if err := binary.Read(buf, binary.LittleEndian, steps); err != nil {
			return err
		}

		track := Track{ID: int(trackIDBuf), Name: string(trackNameBuf[:])}
		for idx, step := range steps {
			track.Steps[idx] = step == 1
		}
		tracks = append(tracks, track)

		header.FileSize -= byte(startTrackSize - buf.Len())
		if header.FileSize <= 0 {
			break
		}
	}

	p.Version = string(bytes.Trim(header.Version[:], "\x00"))
	p.Tempo = header.Tempo
	p.Tracks = tracks

	return nil
}

// String return a representation of pattern
func (p *Pattern) String() string {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, "Saved with HW Version: %s\n", p.Version)
	fmt.Fprintf(&buf, "Tempo: %g\n", p.Tempo)

	for _, track := range p.Tracks {
		fmt.Fprintf(&buf, "(%d) %s\t", track.ID, track.Name)

		fmt.Fprintf(&buf, "|")
		for i := 0; i < measureSteps; i++ {

			for _, step := range track.Steps[i*measureSteps : i*measureSteps+measureSteps] {
				if step {
					fmt.Fprintf(&buf, "%s", "x")
				} else {
					fmt.Fprintf(&buf, "%s", "-")
				}
			}
			fmt.Fprintf(&buf, "|")
		}

		fmt.Fprintln(&buf)
	}

	return buf.String()
}
