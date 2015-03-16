package drum

import (
	"bytes"
	"errors"
	"io"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Write out a string where the total output size is fixed. The output will either
// be truncated or filled with zero/nulls.
func writeFixedNullTermString(spliceFile io.Writer, value string, length int) error {
	var versionBuffer bytes.Buffer
	n, _ := versionBuffer.WriteString(value[:min(length, len(value))])
	if n < length {
		versionBuffer.Write(make([]byte, length-n))
	}
	spliceFile.Write(versionBuffer.Bytes())

	return nil
}

// Read in a fixed length string where a string shorter than given length will be
// inficated by zero/null termination
func readFixedNullTermString(spliceFile io.Reader, length int) (string, error) {
	raw := make([]byte, length)
	n, err := spliceFile.Read(raw)
	if err != nil {
		return "", err
	}
	if n != length {
		return "", errors.New("Could not read string, insufficient data")
	}

	// Find first occurence of null(0) and use it to slice the string
	n = bytes.Index(raw, []byte{0})
	if n < 0 {
		n = length
	}
	return string(raw[:n]), nil
}
