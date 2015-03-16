package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

var tokenSplice = []byte{'S', 'P', 'L', 'I', 'C', 'E'}

// FileHeader holds fixed header info for SPLICE file.
type FileHeader struct {
	fileType [14]byte
	version  [32]uint8
	tempo    float32
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	FileHeader
	tracks []Info
}

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (p *Pattern, err error) {
	p = &Pattern{}
	r, err := os.Open(path)
	if err != nil {
		return
	}
	defer r.Close()

	// FIXME there should be a way here to read the Header in one go
	if err = binary.Read(r, binary.LittleEndian, &p.fileType); err != nil {
		fmt.Println("Failed to read SPLICE file type")
		return
	}

	if bytes.Compare(p.fileType[:5], tokenSplice) == 0 {
		fmt.Println("Header does not match SPLICE file type")
		return
	}

	if err = binary.Read(r, binary.LittleEndian, &p.version); err != nil {
		fmt.Println("Failed to read version")
		return
	}

	if err = binary.Read(r, binary.LittleEndian, &p.tempo); err != nil {
		fmt.Println("Failed to read tempo")
		return
	}

	DecodeBeat(r, p)

	return
}

// String creates a formatted string of the pattern.
func (p *Pattern) String() string {
	var buff bytes.Buffer

	i := bytes.IndexByte(p.version[:], 0)
	buff.WriteString(fmt.Sprint("Saved with HW Version: ", string(p.version[:i]), "\n"))
	buff.WriteString(fmt.Sprintf("Tempo: %.3g\n", p.tempo))

	for _, drum := range p.tracks {
		buff.WriteString(drum.String())
	}

	return buff.String()
}
