package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

// EncodeFile encodes a given pattern to a drum machine file at the provided path.
// If a file already exists at the path, it will be overwritten.
//
// See decoder.go for information on the layout of the file.
func EncodeFile(pattern *Pattern, path string) error {
	// Create our splice file (and defer close).
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write the pattern file into a byte buffer first,
	// so we can count the length.
	var buf bytes.Buffer
	err = pattern.Write(&buf)
	if err != nil {
		return err
	}

	// Write the "SPLICE" header.
	var n int
	header := "SPLICE"
	n, err = f.WriteString(header)
	if err != nil {
		return err
	}
	if n != len(header) {
		return fmt.Errorf("wrote %d bytes, expected to write %d", n, len(header))
	}

	// Write the length as 8-bytes.
	err = binary.Write(f, binary.BigEndian, uint64(buf.Len()))
	if err != nil {
		return err
	}

	// Write the pattern itself (including tracks).
	n, err = f.Write(buf.Bytes())
	if err != nil {
		return err
	}
	if n != buf.Len() {
		return fmt.Errorf("wrote %d bytes, expected to write %d", n, buf.Len())
	}

	return nil
}
