package drum

import (
	"encoding/binary"
	"io"
	"os"
)

// EncodeFile encodes a pattern into a the drum machine file.
func EncodeFile(pattern *Pattern, path string) error {
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = EncodePattern(pattern, out)
	if err != nil {
		return err
	}

	return nil
}

// dataLength calculates the binary data size for the given pattern.
func dataLength(pattern *Pattern) uint8 {
	var size uint8

	size = 32 + // encoder version
		4 // tempo

	for _, track := range pattern.Tracks {
		size += 4 + // track ID
			1 + // name length
			uint8(len(track.Name)) + // name
			16 // steps
	}

	return size
}

// EncodePattern encodes the given pattern into a binary writer.
// It returns the number of bytes written to out, and an eventual error.
func EncodePattern(pattern *Pattern, out io.Writer) (int, error) {
	numBytes, n := 0, 0
	var err error

	// write header
	if n, err = out.Write([]byte(magicHeader)); err != nil {
		return numBytes + n, err
	}
	numBytes += n

	// write data size
	s := dataLength(pattern)
	if n, err = out.Write([]byte{s}); err != nil {
		return numBytes + n, err
	}
	numBytes += n

	// write encoder version
	var ver [32]byte
	copy(ver[:], pattern.HWVersion)
	if n, err = out.Write(ver[:]); err != nil {
		return numBytes + n, err
	}
	numBytes += n

	// write tempo
	if err = binary.Write(out, binary.LittleEndian, pattern.Tempo); err != nil {
		return numBytes, err
	}
	numBytes += 4 // 32 bits

	// write each track
	for _, track := range pattern.Tracks {
		// track ID
		if err = binary.Write(out, binary.LittleEndian, track.ID); err != nil {
			return numBytes, err
		}
		numBytes += 4 // 32 bits

		// name length
		if n, err = out.Write([]byte{byte(len(track.Name))}); err != nil {
			return numBytes + n, err
		}
		numBytes += n

		// name
		if n, err = out.Write([]byte(track.Name)); err != nil {
			return numBytes + n, err
		}
		numBytes += n

		// steps
		// name
		if n, err = out.Write(track.Steps[:]); err != nil {
			return numBytes + n, err
		}
		numBytes += n
	}

	return numBytes, nil
}
