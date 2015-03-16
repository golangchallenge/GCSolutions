package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"text/template"
)

const (
	header = "SPLICE"
)

var byteHeader = []byte(header)
var ticksStrIndices = [16]byte{1, 2, 3, 4, 6, 7, 8, 9, 11, 12, 13, 14, 16, 17, 18, 19}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version string
	Tempo   float32
	Tracks  []*Track
}

const tmpl = `Saved with HW Version: {{.Version}}
Tempo: {{.Tempo}}
{{with .Tracks}}{{range .}}({{.Id}}) {{.Name}}` + "\t" + `{{.Steps}}
{{end}}{{end}}`

var tt = template.Must(template.New("splice").Parse(tmpl))

func (p *Pattern) String() string {
	buf := new(bytes.Buffer)

	err := tt.Execute(buf, p)
	if err != nil {
		panic(err)
	}

	return buf.String()
}

type Ticks uint16

func (s Ticks) String() string {
	res := []byte("|----|----|----|----|")

	mask := Ticks(1)
	for i := 15; i >= 0; i-- {
		if mask&s > 0 {
			res[ticksStrIndices[i]] = 'x'
		}
		mask = mask << 1
	}
	return string(res)
}

// Track in a pattern
type Track struct {
	Name  string
	Id    int
	Steps Ticks
}

// Encoding: <id:byte><name length:uint32 big-endian><name:[]byte><steps:[16]byte>
func decodeTrack(r io.Reader, buf []byte) (*Track, error) {
	t := &Track{}

	b := buf[:1]
	if _, err := io.ReadFull(r, b); err != nil {
		return nil, err
	}

	t.Id = int(b[0])

	var l uint32
	if err := binary.Read(r, binary.BigEndian, &l); err != nil {
		return nil, err
	}

	nameLength := int(l)
	if nameLength > len(buf) {
		b = make([]byte, nameLength)
	} else {
		b = buf[:nameLength]
	}

	if _, err := io.ReadFull(r, b); err != nil {
		return nil, err
	}
	t.Name = string(b)

	b = buf[:16]
	if _, err := io.ReadFull(r, b); err != nil {
		return nil, err
	}

	mask := Ticks(1)
	for i := 15; i >= 0; i-- {
		if b[i] == 1 {
			t.Steps = t.Steps | mask
		}
		mask = mask << 1
	}
	return t, nil
}

// Encoding: <length of pattern: uint64 big-endian><name: [32]byte zero end><tempo: float32 little-endian>
// Returns the pattern and the length in bytes of the tracks section
func decodePatternPreamble(r io.Reader, buf []byte) (*Pattern, int64, error) {
	p := &Pattern{}

	var v int64
	if err := binary.Read(r, binary.BigEndian, &v); err != nil {
		return nil, 0, err
	}

	b := buf[:32]
	if _, err := io.ReadFull(r, b); err != nil {
		return nil, 0, err
	}

	end := 0
	for b[end] > 0 {
		end++
	}
	p.Version = string(b[:end])

	var tempo float32
	if err := binary.Read(r, binary.LittleEndian, &tempo); err != nil {
		return nil, 0, err
	}
	p.Tempo = tempo

	return p, v - 36, nil
}

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	var bufMem [1024]byte
	buf := bufMem[:]

	fr, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fr.Close()

	b := buf[:len(byteHeader)]
	if _, err := io.ReadFull(fr, b); err != nil {
		return nil, err
	}

	if !bytes.Equal(b, byteHeader) {
		return nil, fmt.Errorf("pattern file %s missing required header", path)
	}

	p, l, err := decodePatternPreamble(fr, buf)
	if err != nil {
		return nil, err
	}
	r := io.LimitReader(fr, l)

	for {
		t, err := decodeTrack(r, buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		p.Tracks = append(p.Tracks, t)
	}

	return p, nil
}
