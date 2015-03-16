package drum

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"os"
)

// Encode encodes the pattern to the file found at the provided path.
func Encode(pat *Pattern, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	buf := bufio.NewWriter(file)
	// Write header
	_, err = buf.WriteString(spliceHeader)
	if err != nil {
		return err
	}

	// Write bytes left
	var total uint64
	total = 32                           // version string
	total += 4                           // tempo
	total += uint64(len(pat.Tracks) * 4) // track ids
	total += uint64(len(pat.Tracks) * 1) // inst lengths
	for _, track := range pat.Tracks {
		total += uint64(len(track.Name)) // inst names
	}
	total += uint64(len(pat.Tracks)) * 16 // steps
	err = binary.Write(buf, binary.BigEndian, total)
	if err != nil {
		return err
	}

	// Write version string with null padding
	buf.WriteString(version)
	buf.Write(bytes.Repeat([]byte{0}, 32-len(version)))

	// Write tempo
	err = binary.Write(buf, binary.LittleEndian, pat.Tempo)
	if err != nil {
		return err
	}

	// Write tracks
	for _, track := range pat.Tracks {
		binary.Write(buf, binary.LittleEndian, track.ID)
		binary.Write(buf, binary.BigEndian, uint8(len(track.Name)))
		buf.WriteString(track.Name)
		for _, step := range track.Steps {
			if step {
				buf.WriteByte(1)
			} else {
				buf.WriteByte(0)
			}
		}
	}

	buf.Flush()
	return nil
}
