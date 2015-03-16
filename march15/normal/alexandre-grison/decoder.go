package drum

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	// Open the file
	f, ferr := os.Open(path)
	defer f.Close()
	if ferr != nil {
		return decodeError("Error reading file: "+path, ferr)
	}

	// skip start of file
	f.Seek(13, 0)

	// check how much of the file we should read to avoid invalid file (see sample 5)
	h, e := readNBytesAsHex(1, f)
	if e != nil {
		return decodeError("Error reading file size: ", e)
	}
	bytesToRead, _ := strconv.ParseInt(h, 16, 32)

	// Read the version
	r, e := readNBytes(1, f)
	if e != nil {
		return decodeError("Error reading version: ", e)
	}
	bytesToRead--
	version := make([]string, 0)
	for r[0] != 0 {
		version = append(version, string(r))
		r, e = readNBytes(1, f)
		if e != nil {
			return decodeError("Error reading version: ", e)
		}
		bytesToRead--
	}

	// Keep track of position to correctly update bytesToRead after
	// we'll read the tempo from the .splice
	lastPosition, _ := f.Seek(0, 1)

	// Go to the tempo location in .splice file
	f.Seek(46, 0)
	var tempo float32
	tempoBytes, e := readNBytes(4, f)
	if e != nil {
		return decodeError("Error reading tempo: ", e)
	}

	buf := bytes.NewReader(tempoBytes)
	binary.Read(buf, binary.LittleEndian, &tempo)

	// Compute how much bytes we skipped between last known position
	// and byte ending the tempo data
	currentPosition, _ := f.Seek(0, 1)
	bytesToRead -= 4 + (currentPosition - lastPosition + 2)

	// Now that we've got the version & tempo, init our Pattern
	pattern := NewPattern(strings.Join(version, ""), tempo)

	// iterates on tracks until we reach our limit to read
	for bytesToRead > 0 {
		// Retrieve the track id
		b, e := readNBytesAsHex(1, f)
		if e != nil {
			return decodeError("Error reading track id: ", e)
		}

		id, _ := strconv.ParseInt(b, 16, 32)
		bytesToRead -= 3
		f.Seek(3, 1) // skip 3 bytes

		// Read the track name and its length
		s, e := readNBytes(1, f)
		if e != nil {
			return decodeError("Error reading track name (length): ", e)
		}
		trackNameLength := readIntFromHex(s)

		name, e := readNBytes(trackNameLength, f)
		if e != nil {
			return decodeError("Error reading track name: ", e)
		}

		bytesToRead -= int64(trackNameLength + 1)

		// Read the sequence
		sequence, e := readNBytes(16, f)
		if e != nil {
			return decodeError("Error reading track sequence: ", e)
		}
		bytesToRead -= 16

		// add the current track to our Pattern
		track := NewTrack(id, string(name), sequence)
		pattern.AddTrack(track)
	}

	// ensure Close
	f.Close()

	return pattern, nil
}

// Struct representing a Track inside a Splice
type Track struct {
	number   int64
	name     string
	sequence string
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
// TODO: implement
type Pattern struct {
	version string
	tempo   float32
	tracks  []Track
}

// decodeError returns a nil Pattern and a custom error
func decodeError(label string, e error) (*Pattern, error) {
	return nil, fmt.Errorf(label+"\n\t%#v", e)
}

// readIntFromHex reads an integer from hexadecimal representation of bytes
func readIntFromHex(data []byte) int {
	x := hex.EncodeToString(data)
	i, _ := strconv.Atoi(x)
	return i
}

// readNBytes reads `n` bytes from the file `f`
func readNBytes(n int, f *os.File) ([]byte, error) {
	bytesToRead := make([]byte, n)
	_, e := f.Read(bytesToRead)
	return bytesToRead, e
}

// readNBytesAsHex reads `n` byte from the file `f` as hexadecimal
func readNBytesAsHex(n int, f *os.File) (string, error) {
	bytes, e := readNBytes(n, f)
	if e != nil {
		return "", e
	} else {
		return hex.EncodeToString(bytes), nil
	}
}
