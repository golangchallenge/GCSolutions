package drum

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
	"strings"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
//
// Returns a descriptive error if the procedure fails at any point.
func DecodeFile(path string) (*Pattern, error) {
	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		return nil, err
	}

	return decode(f)
}

func decode(f io.Reader) (*Pattern, error) {
	// verify header, which gives us the expected length of content-body
	expectedBytes, err := verifyHeader(f)
	if err != nil {
		return nil, err
	}
	// read preamble, which gives us basic pattern metadata
	p, err := extractPreamble(f)
	if err != nil {
		return p, err
	}
	// extract all tracks, bearing in mind expected EOD
	remainingBytes := expectedBytes - 36
	p.Tracks, err = extractTracks(f, remainingBytes)

	return p, err
}

/*
	Reads the file header to verify this is a Splice file, and extracts file
	metadata such as version of Splice that created the file.

	Will return an error if the file headers could not be verified as a
	well-formed Splice file.

	The file header consists of the following binary structure:

		| Field   | Bytes | Type    | Notes                     |
		| ------- | ----- | ------- | ------------------------- |
		| ident		|     6 | []byte  | should match "SPLICE"     |
		| size    |     8 | uint64  |                           |

	Returns the expected byte length of the data body within the file as reported
	by the headers. This information should be used as the source of truth, and
	no attempt should be made to read data beyond that length from the file.
*/
func verifyHeader(f io.Reader) (expectedBytes uint64, err error) {
	header := struct {
		Ident    [6]byte
		BodySize uint64 // reported length of the data body within file
	}{}
	readErr := binary.Read(f, binary.BigEndian, &header)
	if readErr != nil {
		return 0, readErr
	}

	// verify filetype identifier
	if header.Ident != [6]byte{'S', 'P', 'L', 'I', 'C', 'E'} {
		return 0, errors.New("File did not have appropriate signature.")
	}

	return header.BodySize, nil
}

/*
	Extracts the "preamble" to a splice file body, which contains general metadata
	about the overall song.

	The preamble consists of the following binary structure:

		| Field   | Bytes | Type    | Notes                     |
		| ------- | ----- | ------- | ------------------------- |
		| version |    32 | []byte  | null padded to length     |
		| tempo   |     4 | float32 | *little endian encoding!* |

	As such, extractPreamble will consume exactly 36 bytes off the buffer reader.

	Returns initialized Pattern struct with all the metadata and an empty track
	array, ready to be populated. Returns an error if parsing failed for any
	reason.
*/
func extractPreamble(f io.Reader) (*Pattern, error) {
	preamble := struct {
		VersionStr [32]byte
		Tempo      float32
	}{}
	readErr := binary.Read(f, binary.LittleEndian, &preamble)
	if readErr != nil {
		return &Pattern{}, readErr
	}

	// trim off C-style null termination
	version := strings.TrimRight(string(preamble.VersionStr[:]), "\x00")

	return &Pattern{Version: version, Tempo: preamble.Tempo}, nil
}

/*
	Extracts all tracks contained in the remainder of content body.

	Should not try to extract beyond the remaining number of expected bytes
	reported in the file header (which needs to be passed to this function).
*/
func extractTracks(f io.Reader, remainingBytes uint64) ([]*Track, error) {
	var tracks []*Track

	for remainingBytes > 0 {
		t, bytesRead, err := extractTrack(f)
		if err != nil {
			return tracks, err
		}

		tracks = append(tracks, t)
		remainingBytes -= bytesRead
	}

	return tracks, nil
}

/*
	Extracts the next track from a splice file body.

	For general info about a track, see the documentation for the Track struct.

	A track consists of the following binary structure:

		| Field    | Bytes | Type    | Notes                                     |
		| -------- | ----- | ------- | ----------------------------------------- |
		| id       |     1 | uint8   |                                           |
		|          |     3 | padding | *appears to be just null padding*         |
		| iLen     |     1 | uint8   | length of following instrument name field |
		| name     |  1..N | uint8   |                                           |
		| beats    |    16 | []uint8 | treated as boolean, can be either 0 or 1  |

	Returns a initialized Track struct with all the metadata and a populated beat
	grid, ready to play some funky tunes.

	Returns the numbers of bytes consumed off the buffer (which is variable), or
	0 in the case of an error (since this number should be not considered reliable
	if there was a failure).

	Returns an error if parsing failed for any reason.
*/
func extractTrack(f io.Reader) (t *Track, bytesConsumed uint64, err error) {
	// need to read the prefix of the track data first, to see length of rest
	prefix := struct {
		ID   uint8
		_    [3]byte // padding
		ILen uint8   // byte length for following instrument name
	}{}
	err = binary.Read(f, binary.BigEndian, &prefix)
	if err != nil {
		return
	}

	// read the instrument name with length from prefix
	instrument := make([]byte, prefix.ILen, prefix.ILen)
	err = binary.Read(f, binary.BigEndian, &instrument)
	if err != nil {
		return
	}

	// read the beat grid
	var beatGrid [16]uint8
	err = binary.Read(f, binary.BigEndian, &beatGrid)
	if err != nil {
		return
	}

	// convert beat grid to local format
	var beats [16]Beat
	for i := range beatGrid {
		beats[i] = beatGrid[i] != 0
	}

	t = &Track{ID: prefix.ID, Name: string(instrument), Beats: beats}
	bytesConsumed = 21 + uint64(prefix.ILen)
	return
}
