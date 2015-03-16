package drum

import (
	"encoding/binary"
	"errors"
	"io"
)

const expectedMagicNumbers = "SPLICE"
const distanceFromSizeToFirstTrack = 0x29
const maxHWVersionLength = 0x20

// readHeader reads header info from fin and populates output
// accordingly.  It also returns the total number of bytes remaining
// to be read for the tracks data.
func readHeader(
	fin io.ReadSeeker,
	output *Pattern,
) (tracksLength int32, err error) {

	// Should have a SPLICE at the beginning
	magicNumbers := make([]byte, 6)
	n, err := fin.Read(magicNumbers)
	if err != nil {
		return
	}
	if n != 6 || string(magicNumbers) != expectedMagicNumbers {
		err = errors.New("Invalid splice file")
		return
	}

	// Seeking to the remaining file size
	_, err = fin.Seek(0xa, 0)
	if err != nil {
		return
	}
	// It seems odd that this would be a big-endian number when the
	// tempo and track numbers are little-endian, but the alternative
	// would be that the tracks couldn't possibly contain more than
	// 255 bytes, so I'll just treat it as a big-endian 32-bit uint.
	// It won't hurt either way for the given tests since the first
	// three bytes are all zeroes in the example files.
	err = binary.Read(fin, binary.BigEndian, &tracksLength)
	if err != nil {
		return
	}
	tracksLength -= distanceFromSizeToFirstTrack

	// HW Version text is nil-terminated
	output.HWVersion = readCString(fin, maxHWVersionLength)

	// Tempo is a 32-bit float, little-endian encoded
	_, err = fin.Seek(0x2e, 0)
	err = binary.Read(fin, binary.LittleEndian, &output.Tempo)

	return
}
