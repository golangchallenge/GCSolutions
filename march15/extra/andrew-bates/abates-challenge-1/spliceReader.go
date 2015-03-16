package drum

import (
	"bytes"
	"encoding/binary"
	"io"
	"os"
)

// Reader reads fields from binary streams and keeps track of any errors
// encountered.  Implementations of Reader should keep track of errors
// encountered while reading and expose the error with the Err function.  If an
// error has occurred, subsequent calls to Next, FixedString and VarString
// should do nothing
//
// Next reads into the provided interface until it's full.  It uses binary.Read
// to populate the destination interface
//
// FixedString reads a fixed length field into a string destination.  The
// length should be in bytes, not characters
//
// VarString reads a variable length string into the destination.  This
// function assumes the first byte of the stream is the length of the string
// and will then read that many bytes.
//
// Err should return the most recent error encountered while reading nor nil if
// no error has occurred
type Reader interface {
	Next(dst interface{})
	FixedString(uint8, *string)
	VarString(*string)
	Err() error
}

type commonReader struct {
	input io.Reader
	err   error
}

func (r *commonReader) Next(dst interface{}) {
	if r.err == nil {
		r.err = binary.Read(r.input, binary.LittleEndian, dst)
	}
}

func (r *commonReader) FixedString(length uint8, dst *string) {
	buffer := make([]byte, length)
	r.Next(buffer)
	trim := bytes.IndexByte(buffer, 0x00)
	if trim > 0 {
		*dst = string(buffer[:trim])
	} else {
		*dst = string(buffer[:])
	}
}

func (r *commonReader) VarString(dst *string) {
	var length uint8
	r.err = binary.Read(r.input, binary.LittleEndian, &length)
	r.FixedString(length, dst)
}

func (r *commonReader) Err() error {
	return r.err
}

type fileReader struct {
	commonReader
	file *os.File
}

func (f *fileReader) Close() error {
	return f.file.Close()
}

func newFileReader(path string) Reader {
	file, err := os.Open(path)
	return &fileReader{
		commonReader{
			input: file,
			err:   err,
		},
		file,
	}
}

func newByteArrayReader(splice []byte) Reader {
	return &commonReader{
		input: bytes.NewReader(splice),
		err:   nil,
	}
}
