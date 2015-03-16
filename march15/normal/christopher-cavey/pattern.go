package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

func NewPattern() *Pattern {
	return &Pattern{
		Version: "1.0-golang",
		Tempo:   120,
	}
}

func (p *Pattern) AddTrack(track *Track) error {
	p.Tracks = append(p.Tracks, track)
	return nil
}

func (p *Pattern) NewTrack(ID int32, name string) (*Track, error) {
	return nil, nil
}

func (p *Pattern) GetTrackByID(ID int32) *Track {
	return nil
}

func (p *Pattern) GetTrackByName(name string) *Track {
	return nil
}

func (p *Pattern) String() string {
	var buffer bytes.Buffer

	fmt.Fprintf(&buffer, "Saved with HW Version: %s\n", p.Version)
	fmt.Fprintf(&buffer, "Tempo: %g", p.Tempo)

	for i, track := range p.Tracks {
		if i < len(p.Tracks) {
			buffer.WriteString("\n")
		}
		buffer.WriteString(track.String())
	}

	buffer.WriteString("\n")

	return buffer.String()
}

func (p *Pattern) Write(path string) error {
	spliceFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer spliceFile.Close()

	n, err := spliceFile.Write(spliceMagic)
	if err != nil {
		return err
	}
	if n != len(spliceMagic) {
		return fmt.Errorf("Could not write %d bytes to %s", len(spliceMagic), path)
	}

	// Skip over dataSize since we won't know until the end
	spliceFile.Seek(8, os.SEEK_CUR)

	writeFixedNullTermString(spliceFile, p.Version, 32)

	binary.Write(spliceFile, binary.LittleEndian, p.Tempo)

	var dataSize int64 = 36 // starting  datasize is 32 byte verion + 4 byte tempo

	for _, track := range p.Tracks {
		n, _ := track.Write(spliceFile)
		dataSize += int64(n)
	}

	// Write this at the end
	spliceFile.Seek(6, os.SEEK_SET)
	binary.Write(spliceFile, binary.BigEndian, dataSize)

	return nil
}
