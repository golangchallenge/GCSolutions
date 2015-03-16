package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
)

// The length in bytes of Pattern and Track attributes
const (
	softwareLen = 10
	versionLen  = 32
	tempoLen    = 4
	idLen       = 1
	sizeLen     = 4
	stepsLen    = 16
)

// A Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Software string
	Name     string
	Version  string
	Tempo    float32
	Tracks
	Size int32
}

// A Track is the high level representation of the
// pattern tracks contained in a .splice file.
type Track struct {
	ID    byte
	Size  int32
	Name  string
	Steps []byte
}

// A Tracks is an array of Track
type Tracks []Track

func (p Pattern) String() string {
	str := fmt.Sprintf("Saved with HW Version: %v\n", strings.Trim(p.Version, "\x00"))
	str = fmt.Sprintf("%vTempo: %v\n", str, p.Tempo)
	for _, t := range p.Tracks {
		str = fmt.Sprintf("%v%v", str, t)
	}

	return str
}

func (t Track) String() string {
	str := fmt.Sprintf("(%v) %v\t", t.ID, t.Name)
	for i, v := range t.Steps {
		if (i % 4) == 0 {
			str = fmt.Sprintf("%v|", str)
		}
		if v == 0 {
			str = fmt.Sprintf("%v-", str)
		} else {
			str = fmt.Sprintf("%vx", str)
		}
	}
	str = fmt.Sprintf("%v|\n", str)
	return str
}

// DecodeFile opens the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (p *Pattern, err error) {
	file, err := os.Open(path)
	if err != nil {
		log.Printf("error opening file %v:%v", path, err)
		return nil, err
	}
	defer file.Close()

	p = new(Pattern)

	p.Load(file)

	return p, nil
}

// Load decodes the provided drum machine file to the pattern.
func (p *Pattern) Load(file *os.File) {
	var wg sync.WaitGroup
	wg.Add(4)

	go func() {
		defer wg.Done()
		//Reading software name
		readString(&p.Software, 0, softwareLen, file)
	}()

	go func() {
		defer wg.Done()
		//Reading software Version
		readString(&p.Version, softwareLen+sizeLen, versionLen, file)
	}()

	go func() {
		defer wg.Done()
		//Reading Pattern Tempo
		readFloat32(&p.Tempo, softwareLen+sizeLen+versionLen, file)
	}()

	go func() {
		defer wg.Done()
		//Reading Pattern Size
		readInt32(&p.Size, softwareLen, file)

		//Reading Track data
		p.Tracks.load(p, file)
	}()

	wg.Wait()
}

func (tracks *Tracks) load(p *Pattern, file *os.File) {
	//Track data offset
	off := int64(softwareLen + sizeLen + versionLen + tempoLen)

	aTrack := []Track{}
	//Reading Tracks
	for off < int64(p.Size) {
		t := Track{}
		off = t.load(off, file)
		aTrack = append(aTrack, t)
	}
	*tracks = aTrack
}

func (t *Track) load(off int64, file *os.File) (n int64) {
	//Reading Track ID
	readOneByte(&t.ID, off, file)
	off = off + idLen

	//Reading Track Size
	readInt32(&t.Size, off, file)
	off = off + sizeLen

	//Reading Track Name
	readString(&t.Name, off, int(t.Size), file)
	off = off + int64(t.Size)

	//Reading Track Steps
	readByte(&t.Steps, off, stepsLen, file)
	off = off + stepsLen

	return off
}

// readByte reads n (length) bytes from file, begging at off
func readByte(b *[]byte, off int64, length int, file *os.File) {
	// The byte array must have a fixed length
	data := make([]byte, length, length)
	_, err := file.ReadAt(data, off)
	if err != nil {
		log.Printf("error reading %d bytes from file %v at offset %d:%v", length, file.Name(), off, err)
		return
	}

	//log.Printf("% x ", data)
	*b = data
}

func readFloat32(f *float32, off int64, file *os.File) {
	b := make([]byte, 4, 4)
	readByte(&b, off, 4, file)
	binary.Read(bytes.NewReader(b), binary.LittleEndian, f)
}

func readInt32(i *int32, off int64, file *os.File) {
	b := make([]byte, 4, 4)
	readByte(&b, off, 4, file)
	*i = int32(binary.BigEndian.Uint32(b))
}

func readString(s *string, off int64, length int, file *os.File) {
	b := make([]byte, length, length)
	readByte(&b, off, length, file)
	*s = string(b[0:])
}

func readOneByte(b *byte, off int64, file *os.File) {
	d := make([]byte, 1, 1)
	readByte(&d, off, 1, file)
	*b = d[0]
}
