package drum

import (
	"encoding/binary"
	"os"
)

// EncodePattern encodes the pattern to binary format and saves it to a file.
// Pretty much the opposite of `DecodeFile`.
func EncodePattern(p *Pattern, path string) error {
	var err error
	var magic []byte = []byte("SPLICE")
	var patLen byte
	var version []byte = []byte(DrumVersion)
	var tempo float32 = p.Tempo

	// First we have to create the file.
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Helper function to write `n` null bytes to the file.
	writeNullBytes := func(n int) error {
		b := make([]byte, n)
		return binary.Write(f, binary.LittleEndian, b)
	}

	// Write the magic number to the head of the file.
	err = binary.Write(f, binary.LittleEndian, magic)
	if err != nil {
		return err
	}

	// Write 7 null bytes.
	err = writeNullBytes(7)
	if err != nil {
		return err
	}

	// Calculate and write the pattern length. Basically: `versionString + tempo + ((id + nameLen + name) * 16)`.
	patLen = byte(len(DrumVersion)) + 4
	for _, t := range p.Tracks {
		patLen += 4 + 1 + byte(len(t.Name)) + 16
	}
	err = binary.Write(f, binary.LittleEndian, patLen)
	if err != nil {
		return err
	}

	// Write the version string.
	err = binary.Write(f, binary.LittleEndian, version)
	if err != nil {
		return err
	}

	// The version string is expected to be a char[32], so we have to fill in some
	// null bytes till we have reached 32 bytes.
	err = writeNullBytes(32 - len(DrumVersion))
	if err != nil {
		return err
	}

	// Write the tempo.
	err = binary.Write(f, binary.LittleEndian, tempo)
	if err != nil {
		return err
	}

	// Loop through each track.
	for _, t := range p.Tracks {
		var nameLen byte = byte(len(t.Name))
		var steps []byte = make([]byte, 16)

		// Write the id, name length and name

		err = binary.Write(f, binary.LittleEndian, t.Id)
		if err != nil {
			return err
		}

		err = binary.Write(f, binary.LittleEndian, nameLen)
		if err != nil {
			return err
		}

		err = binary.Write(f, binary.LittleEndian, []byte(t.Name))
		if err != nil {
			return err
		}

		// Write the step bytes (`0` for false and `1` for true).
		for i, s := range t.Steps {
			if s {
				steps[i] = 1
			} else {
				steps[i] = 0
			}
		}
		err = binary.Write(f, binary.LittleEndian, steps)
	}

	return nil
}
