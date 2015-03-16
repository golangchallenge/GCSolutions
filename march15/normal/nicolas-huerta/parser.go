package drum

import (
	"bytes"
	"encoding/binary"
	"math"
	"strings"
)

// ParseFloat32LE reads a byteSlice and parses the bytes as a 32 bit little endian float and returns a float32
func ParseFloat32LE(byteSlice []byte) (float32, error) {
	var floatValue float32
	byteSlice = byteSlice[:binary.Size(floatValue)]
	buf := bytes.NewReader(byteSlice)
	err := binary.Read(buf, binary.LittleEndian, &floatValue)
	return floatValue, err
}

// ParseInt32BE reads a byteSlice and parses the bytes as a 32 bit big endian integer and returns an int32
func ParseInt32BE(byteSlice []byte) (int32, error) {
	var intValue int32
	byteSlice = byteSlice[:binary.Size(intValue)]
	buf := bytes.NewReader(byteSlice)
	err := binary.Read(buf, binary.BigEndian, &intValue)
	return intValue, err
}

// ParseUInt8BE reads a byteSlice and parses the bytes as a 8 bit big endian unsigned integer and returns an uint8
func ParseUInt8BE(byteSlice []byte) (uint8, error) {
	var value byte
	byteSlice = byteSlice[:binary.Size(value)]
	buf := bytes.NewReader(byteSlice)
	err := binary.Read(buf, binary.BigEndian, &value)
	return value, err
}

// ParseCString reads a byteSlice until the first NUL character and returns a string
func ParseCString(byteSlice []byte) string {
	strIndex := strings.Index(string(byteSlice), "\x00")
	stringValue := string(byteSlice[:strIndex])
	return stringValue
}

// ParseString reads at most length bytes from a byteSlice and returns a string
func ParseString(byteSlice []byte, length int) string {
	return string(byteSlice[:length])
}

// ParseTrackPatternAsString decodes a sequence of bytes representing
// track pattern step data in groups of 4 consecutive bytes
// example output looks like this: |x-x-|x-x-|x-x-|x-x-|
func ParseTrackPatternAsString(trackPatternBytes []byte) string {
	trackPatternStr := ""

	for key, value := range trackPatternBytes[:kTrackPatternLength] {
		if math.Mod(float64(key), 4) == 0 {
			trackPatternStr += "|"
		}
		if value == 0 {
			trackPatternStr += "-"
		}
		if value == 1 {
			trackPatternStr += "x"
		}
	}
	trackPatternStr += "|\n"
	return trackPatternStr
}
