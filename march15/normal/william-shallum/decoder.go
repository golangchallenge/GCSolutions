package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"text/template"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (pattern *Pattern, err error) {
	defer func() {
		x := recover()
		if x != nil {
			pattern = nil
			panicErr, ok := x.(error)
			if !ok {
				err = errors.New("unknown error")
			} else {
				err = panicErr
			}
		}
	}()
	return decodeFileInternal(path), nil
}
const signatureSize uint = 12

// decodeFileInternal actually decodes the file. 
// If an error occurs it will panic with the error and close the file.
// DecodeFile will recover from the panic and return the error to its caller.
func decodeFileInternal(path string) *Pattern {
	var le binary.ByteOrder = binary.LittleEndian
	var be binary.ByteOrder = binary.BigEndian

	f, err := os.Open(path)
	if (err != nil) {
		panic(err)
	}
	defer f.Close()

	fi, err := os.Stat(path)
	if (err != nil) {
		panic(err)
	}

	sig := make([]uint8, signatureSize)
	readOrPanic(f, le, sig)
	verifySignature(sig)

	var dataLength uint16
	readOrPanic(f, be, &dataLength)
	verifyDataLength(fi.Size(), dataLength)

	hwVersion := make([]uint8, 32)
	readOrPanic(f, le, hwVersion)

	var tempo float32
	readOrPanic(f, le, &tempo)

	currentOffset := ftell(f)

	var endOfDataOffset = int64(uint(dataLength) + signatureSize)
	tracks := make([]PatternTrack, 0, 16)
	for currentOffset < endOfDataOffset {
		var track PatternTrack

		var trackID uint32
		readOrPanic(f, le, &trackID)

		var trackNameLen uint8
		readOrPanic(f, le, &trackNameLen)

		trackNameBuf := make([]uint8, trackNameLen)
		readOrPanic(f, le, trackNameBuf)

		stepsBuf := make([]uint8, 16)
		readOrPanic(f, le, stepsBuf)

		track.ID = trackID
		track.Name = bytesToString(trackNameBuf)
		for i := range stepsBuf {
			track.Steps[i] = (stepsBuf[i] != 0)
		}
		tracks = append(tracks, track)

		currentOffset = ftell(f)
	}

	p := &Pattern{}
	p.HwVersion = bytesToString(hwVersion)
	p.Tempo = tempo
	p.Tracks = tracks
	return p
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file. It contains multiple tracks.
type Pattern struct {
	HwVersion string
	Tempo     float32
	Tracks    []PatternTrack
}

// PatternTrack is a track within the .splice file.
type PatternTrack struct {
	ID    uint32
	Name  string
	Steps [16]bool // true if a sound is emitted at that step.
}

var patternTemplate = template.Must(template.New("pattern").Parse(
	"Saved with HW Version: {{.HwVersion}}\n" +
		"Tempo: {{.Tempo}}\n" +
		"{{range .Tracks}}" +
		"({{.ID}}) {{.Name}}\t{{.StepsAsString}}\n" +
		"{{end}}"))

// Formats the pattern as a nice string.
func (p *Pattern) String() string {
	var buf bytes.Buffer
	patternTemplate.Execute(&buf, p)
	return buf.String()
}

// StepsAsString returns the steps of a track as a formatted string.
func (t *PatternTrack) StepsAsString() string {
	var buf bytes.Buffer
	for i := range t.Steps {
		if i%4 == 0 {
			buf.WriteString("|")
		}
		if t.Steps[i] {
			buf.WriteString("x")
		} else {
			buf.WriteString("-")
		}
	}
	buf.WriteString("|")
	return buf.String()
}

// bytesToString converts the bytes in the buffer to a string. 
// The buffer is treated as nul terminated, only the bytes up to 
// the first nul will be converted.
func bytesToString(b []byte) string {
	nulIndex := 0
	nulFound := false
	for nulIndex = range b {
		if b[nulIndex] == 0 {
			nulFound = true
			break
		}
	}
	if nulFound {
		return string(b[:nulIndex])
	}
	return string(b)
}

// readOrPanic wraps binary.Read and panics on error.
// This saves 3 lines per call in decodeFileInternal.
func readOrPanic(r io.Reader, order binary.ByteOrder, data interface{}) {
	err := binary.Read(r, order, data)
	if err != nil {
		panic(err)
	}
}

// verifySignature verifies the signature: "SPLICE" followed by 6 NUL bytes.
// It panics if the signature cannot be verified.
func verifySignature(sig []byte) {
	var goodSig = [12]uint8{0x53, 0x50, 0x4c, 0x49, 0x43, 0x45,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	if len(sig) != 12 {
		panic(errors.New("bad signature length"))
	}
	for i := range sig {
		if sig[i] != goodSig[i] {
			panic(errors.New("bad signature"))
		}
	}
}

// verifyDataLength checks that the data length in the header can fit into 
// the file.
func verifyDataLength(fileSize int64, dataLength uint16) {
	if fileSize < int64(uint(dataLength) + signatureSize) {
		panic("file too small for data length")
	}
}

// ftell gets the current offset in a file or panics on error.
func ftell(f *os.File) int64 {
	currentOffset, err := f.Seek(0, os.SEEK_CUR)
	if err != nil {
		panic(err)
	}
	return currentOffset
}
