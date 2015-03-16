package drum

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
)

const nullByte = 0x00

const (
	dashSymbol  = 0x00
	crossSymbol = 0x01
)

const (
	bytePositionContainingFileLenData = 13
	byteCountOffsetToFileLen          = 14

	formatByteStartPosition = 0
	formatByteEndPosition   = 14

	versionByteStartPosition = 14
	versionByteEndPosition   = 46

	tempoByteStartPosition = 46
	tempoByteEndPosition   = 50

	tracksByteStartPosition = 50
)

type decoder interface {
	DecodeByteValuesToString() string
}

// formatByteSlice, versionByteSlice, tempoByteSlice,
// and trackBytesSlice implement the decoder interface.
type formatByteSlice struct{ bytes []byte }

type versionByteSlice struct{ bytes []byte }

type tempoByteSlice struct{ bytes []byte }

type trackBytesSlice struct{ bytes []byte }

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct{ pattern string }

func (p *Pattern) String() string {
	return fmt.Sprint(p.pattern)
}

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	var data []string

	for _, byteSlice := range getByteSlicesFromFile(path) {
		data = append(data, byteSlice.DecodeByteValuesToString())
	}

	p := &Pattern{
		pattern: fmt.Sprintf(
			`Saved with HW Version: %s
Tempo: %s
%s`,
			data[0],
			data[1],
			data[2]),
	}

	return p, nil
}

func getByteSlicesFromFile(path string) []decoder {
	totalFileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal("error reading file: ", err)
	}

	// Don't bother executing the rest of the program if not SPLICE format.
	format := &formatByteSlice{
		bytes: totalFileBytes[formatByteStartPosition:formatByteEndPosition],
	}
	if fFmt := format.DecodeByteValuesToString(); fFmt != "SPLICE" {
		log.Fatal("file decoding process terminated: not splice file format, got ", fFmt)
	}

	// Break file bytes into byte slices to decode relevant data.
	return []decoder{
		&versionByteSlice{
			bytes: totalFileBytes[versionByteStartPosition:versionByteEndPosition],
		},
		&tempoByteSlice{
			bytes: totalFileBytes[tempoByteStartPosition:tempoByteEndPosition],
		},
		&trackBytesSlice{
			// FileLenData returns the original number of bytes after the byte holding the FileLenData value. Add offset bytes to get the correct original file size.
			bytes: totalFileBytes[tracksByteStartPosition : totalFileBytes[bytePositionContainingFileLenData]+byteCountOffsetToFileLen],
		},
	}
}

func (f formatByteSlice) DecodeByteValuesToString() string {
	return decodeBytesUntilNullByte(f.bytes)
}

func (v versionByteSlice) DecodeByteValuesToString() string {
	return decodeBytesUntilNullByte(v.bytes)
}

func decodeBytesUntilNullByte(sl []byte) string {
	// Create a buffer and read the current byte from the given slice into it.
	buf := new(bytes.Buffer)
	r := bufio.NewReader(bytes.NewBuffer(sl))

	// Read every subsequent byte into the buffer.
	// Null byte will cause the loop to exit.
	for {
		if b := readByte(r); b != nullByte {
			writeByte(buf, b)
		} else {
			break
		}
	}
	return buf.String()
}

func (t tempoByteSlice) DecodeByteValuesToString() string {
	var tempo float32
	buf := bytes.NewReader(t.bytes)
	err := binary.Read(buf, binary.LittleEndian, &tempo)
	if err != nil {
		fmt.Println("binary.Read failed: ", err)
	}

	return fmt.Sprint(tempo)
}

func (t trackBytesSlice) DecodeByteValuesToString() string {
	buf := new(bytes.Buffer)
	r := bufio.NewReader(bytes.NewBuffer(t.bytes))

	for {
		writeTrackIDStringToBuf(buf, r)
		writeTrackNameBytesToBuf(buf, r)
		writeTrackStepPatternBytesToBuf(buf, r)
		// Check if EOF without advancing reader.
		_, err := r.Peek(1)
		if err != nil {
			break
		}
	}

	return buf.String()
}

// writeTrackIDStringToBuf writes to buffer a single ID string without name and steps, eg. (1) .
func writeTrackIDStringToBuf(buf *bytes.Buffer, r *bufio.Reader) {
	ID := fmt.Sprintf("(%v) ", readByte(r))

	// ID byte is followed by 3 null bytes. Consume all 3 bytes.
	for i := 0; i < 3; i++ {
		readByte(r)
	}

	_, err := buf.WriteString(ID)
	if err != nil {
		fmt.Println("error writing string to buffer ", err)
	}
}

// writeTrackNameBytesToBuf writes bytes of a single track name to buffer, eg. kick\t.
func writeTrackNameBytesToBuf(buf *bytes.Buffer, r *bufio.Reader) {
	nameByteLen := readByte(r)
	for i := 0; i < int(nameByteLen); i++ {
		writeByte(buf, readByte(r))
	}
	writeByte(buf, '\t')
}

// writeTrackStepPatternBytesToBuf writes bytes of a single track name to buffer, eg. |x---|----|x---|----|\n.
func writeTrackStepPatternBytesToBuf(buf *bytes.Buffer, r *bufio.Reader) {
	for i := 0; i < 16; i++ {
		if i%4 == 0 {
			writeByte(buf, '|')
		}
		writeByte(buf, decodeStepPattern(r))
	}
	writeByte(buf, '|')
	writeByte(buf, '\n')
}

func decodeStepPattern(r *bufio.Reader) byte {
	b := readByte(r)

	switch b {
	case dashSymbol:
		return '-'
	case crossSymbol:
		return 'x'
	default:
		log.Fatalf("file corrupted: step pattern %v not recognised by SPLICE decoder)", b)
	}

	return 0
}

func readByte(r *bufio.Reader) byte {
	b, err := r.ReadByte()
	if err != nil {
		fmt.Println("error reading byte: ", err)
	}
	return b
}

func writeByte(buf *bytes.Buffer, b byte) {
	err := buf.WriteByte(b)
	if err != nil {
		fmt.Println("error writing byte: ", err)
	}
}
