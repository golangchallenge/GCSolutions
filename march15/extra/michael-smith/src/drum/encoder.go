package drum

import (
	"bytes"
	"encoding/binary"
	"io"
	"os"
)

// EncodeFile will write a Pattern to a file
func EncodeFile(path string, p *Pattern) error {
	// create new file
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// encode and write to the file
	if err := Encode(p, f); err != nil {
		return err
	}

	return nil
}

// Encode will write a Pattern to a writer
func Encode(pattern *Pattern, writer io.Writer) error {
	// First, write the pattern & track data to a buffer.
	// We will need to get its length before writing the file header
	buff := new(bytes.Buffer)

	var ph patternHeader
	copy(ph.HwVers[0:], []byte(pattern.Version))
	ph.Tempo = pattern.Tempo

	if err := binary.Write(buff, binary.LittleEndian, ph); err != nil {
		return err
	}

	for _, track := range pattern.Tracks {
		// write ID
		if err := binary.Write(buff, binary.LittleEndian, track.ID); err != nil {
			return err
		}

		// write nameLen
		nameLen := uint8(len(track.Name))
		if err := binary.Write(buff, binary.LittleEndian, nameLen); err != nil {
			return err
		}

		// write name
		if _, err := buff.WriteString(track.Name); err != nil {
			return err
		}

		// write steps
		steps := make([]byte, 16)
		for i, b := range track.Steps {
			if b {
				steps[i] = 1
			} else {
				steps[i] = 0
			}
		}
		if _, err := buff.Write(steps); err != nil {
			return err
		}
	}

	// Write file header
	var fh fileHeader
	copy(fh.Signature[0:], []byte(fileSignature))
	fh.Length = uint64(buff.Len())
	if err := binary.Write(writer, binary.BigEndian, fh); err != nil {
		return err
	}

	// Now, write the pattern & track data buffer after the file header
	if _, err := io.Copy(writer, buff); err != nil {
		return err
	}

	return nil
}
