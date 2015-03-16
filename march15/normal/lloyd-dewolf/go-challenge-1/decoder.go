package drum

///package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	//	"github.com/davecgh/go-spew/spew"
	"io"
	"os"
	//	"path"
	"unicode"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func getPrintableString(dirty string) string {
	var b bytes.Buffer
	for _, by := range dirty {
		if unicode.IsPrint(rune(by)) {
			b.WriteString(string(by))
		}
	}
	return b.String()
}

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
// TODO: implement
func DecodeFile(fullFileName string) (*Pattern, error) {

	// r implments io.ReadWriteCloser
	r, e := os.Open(fullFileName)
	check(e)
	defer r.Close()
	fi, e := r.Stat()
	check(e)
	szFile := fi.Size()

	var p Pattern

	// Read header from file
	header := make([]byte, 6)
	_, e = r.Read(header)
	check(e)
	if string(header) != "SPLICE" {
		e = errors.New("File missing header")
	}
	check(e)

	// Read version
	const VPos int64 = 14
	_, e = r.Seek(VPos, 0)
	check(e)
	b1 := make([]byte, 30)
	_, e = r.Read(b1)
	check(e)
	p.Version = getPrintableString(string(b1))

	// Read tempo
	const TPos int64 = 46
	_, e = r.Seek(TPos, 0)
	check(e)
	binary.Read(r, binary.LittleEndian, &p.Tempo)

	// TODO: current implimentation assumes at least two sounds.
	//       Move to a single loop.

	// Read 1st sound
	// 	Read s.ID
	var s Sound
	binary.Read(r, binary.LittleEndian, &s.ID)

	// 	Read s.Name
	dirtySoundName := make([]byte, 20)
	_, e = r.Read(dirtySoundName)
	check(e)
	s.Name = getPrintableString(string(dirtySoundName))

	// 	Read s.Steps
	// TODO: all these literals should be constants
	namePos := TPos + 4 + 1 + 4
	l := len(s.Name)
	stepsPos := namePos + int64(l)
	_, e = r.Seek(stepsPos, 0)
	check(e)
	steps := make([]byte, 16)
	n, e := r.Read(steps)
	check(e)
	s.Steps = steps
	endSoundPos := namePos + int64(l) + int64(n)

	// 	Add sound to pattern
	p.Sounds = append(p.Sounds, s)

	for i := 0; endSoundPos < szFile; i++ {
		// 0.708-alpha is known to have to have garbage
		if p.Version == "0.708-alpha" {
			headerAgain := make([]byte, 6)
			_, e = r.Read(headerAgain)
			if string(headerAgain) == "SPLICE" {
				break // we're done
			} else {
				_, e = r.Seek(endSoundPos, 0)
			}
		}
		// Read another sound
		s = Sound{0, "", nil}
		// 	Read s.ID
		e = binary.Read(r, binary.LittleEndian, &s.ID)
		check(e)
		namePos = endSoundPos + 1 + 4
		// 	Read s.Name
		dirtySoundName = make([]byte, 20)
		_, e = r.Read(dirtySoundName)
		if e == io.EOF {
			break
		}
		check(e)
		s.Name = getPrintableString(string(dirtySoundName))

		// 	Read s.Steps
		l = len(s.Name)
		stepsPos = endSoundPos + 4 + 1 + int64(l)
		_, e = r.Seek(stepsPos, 0)
		check(e)
		steps = make([]byte, 16)
		n, e = r.Read(steps)
		check(e)
		s.Steps = steps
		endSoundPos = namePos + int64(l) + int64(n)

		// 	Add sound to pattern
		p.Sounds = append(p.Sounds, s)
	}

	return &p, e
}

// Sound is a track of a drum machine.
// TODO: maybe Sound shouldn't be an exported type.
type Sound struct {
	ID    int32
	Name  string
	Steps []byte
}

func (s *Sound) String() string {
	var b bytes.Buffer
	b.WriteString("(" + fmt.Sprintf("%v", s.ID) + ") ")
	b.WriteString(s.Name + "\t" + "|")
	for i, step := range s.Steps {
		i++
		if step == 1 {
			b.WriteString("x")
		} else {
			b.WriteString("-")
		}

		if (i % 4) == 0 {
			b.WriteString("|")
		}
	}
	return b.String()
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version string
	Tempo   float32
	Sounds  []Sound
}

func (p *Pattern) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("Saved with HW Version: " + p.Version + "\n")
	buffer.WriteString("Tempo: " + fmt.Sprintf("%g", p.Tempo) + "\n")
	for _, sound := range p.Sounds {
		buffer.WriteString(fmt.Sprint(&sound) + "\n")
	}
	return buffer.String()
}

/*
func main() {
	//	filenames := []string{"pattern_1.splice", "pattern_2.splice", "pattern_3.splice", "pattern_4.splice", "pattern_5.splice"}
	filenames := []string{"pattern_5.splice"}

	for _, filename := range filenames {
		fmt.Println(filename)
		decoded, err := DecodeFile(path.Join("fixtures", filename))
		check(err)
		if err == nil {
			fmt.Println(decoded)
		}
	}
}
*/
