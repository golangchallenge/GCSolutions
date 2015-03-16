package drum

import (
        "bytes"
        "encoding/binary"
        "fmt"
        "io"
        "io/ioutil"
        "math"
)

const (
        stepSize  = 16 // Number of steps
        headerOff = 36 // Size of Header without FileID
        trackWN   = 21 // Size of Track without Name
)

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
        // header
        FileID   []byte
        DataSize uint8
        Version  []byte
        Tempo    float32

        // tracks
        Tracks []track
}

type track struct {
        trackID  uint8
        res      []byte
        nameSize uint8
        name     []byte
        steps    []byte
}

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
        pat := &Pattern{}
        b, err := ioutil.ReadFile(path)
        if err != nil {
                return nil, err
        }
        buf, err := pat.decodeHead(b)
        if err != nil {
                return nil, err
        }
        // file is not long enough, must be corrupt
        if int(pat.DataSize-headerOff) > buf.Len() {
                return nil, fmt.Errorf("drum: DecodeFile: The file '%v' is corrupt", path)
        }
        err = pat.decodeTracksAll(buf)
        if err != nil {
                return nil, err
        }
        return pat, nil
}

func (p *Pattern) String() string {
        var buffer bytes.Buffer
        buffer.WriteString("Saved with HW Version: " + string(p.Version))
        buffer.WriteString("\nTempo: " + fmt.Sprint(p.Tempo))
        for _, vali := range p.Tracks {
                buffer.WriteString("\n(" + fmt.Sprint(vali.trackID) + ") ")
                buffer.WriteString(string(vali.name) + "\t")
                for j, valj := range vali.steps {
                        if math.Mod(float64(j), 4) == 0 {
                                buffer.WriteString("|")
                        }
                        if valj == 0x00 {
                                buffer.WriteString("-")
                        } else {
                                buffer.WriteString("x")
                        }
                }
                buffer.WriteString("|")
        }
        buffer.WriteString("\n")
        return buffer.String()
}

func (p *Pattern) decodeHead(b []byte) (*bytes.Reader, error) {
        buf := bytes.NewReader(b)
        var fileID [13]byte
        err := binary.Read(buf, binary.LittleEndian, &fileID)
        if err != nil {
                return nil, err
        }
        p.FileID = sliceToStr(fileID[:])
        err = binary.Read(buf, binary.LittleEndian, &p.DataSize)
        if err != nil {
                return nil, err
        }
        var version [32]byte
        err = binary.Read(buf, binary.LittleEndian, &version)
        if err != nil {
                return nil, err
        }
        p.Version = sliceToStr(version[:])
        err = binary.Read(buf, binary.LittleEndian, &p.Tempo)
        if err != nil {
                return nil, err
        }
        return buf, nil
}

func (p *Pattern) decodeTracksAll(buf *bytes.Reader) error {
        var total int
        for {
                n, err := p.decodeTrack(buf)
                if err != nil {
                        return err
                }
                total += n
                if total == int(p.DataSize-headerOff) {
                        break
                }
        }
        return nil
}

func (p *Pattern) decodeTrack(buf *bytes.Reader) (int, error) {
        var b []byte
        b = make([]byte, 5)
        if _, err := io.ReadFull(buf, b); err != nil {
                return 0, err
        }
        var track track
        track.trackID = b[0]
        track.res = b[1:3]
        track.nameSize = b[4]
        b = nil
        b = make([]byte, track.nameSize+stepSize)
        if _, err := io.ReadFull(buf, b); err != nil {
                return 0, err
        }
        track.name = b[:track.nameSize]
        track.steps = b[track.nameSize : track.nameSize+stepSize]
        p.Tracks = append(p.Tracks, track)
        return int(track.nameSize + trackWN), nil
}

// find the first 0x00 and cut the rest
func sliceToStr(b []byte) []byte {
        buf := bytes.NewReader(b)
        var n []byte
        n = make([]byte, 1)
        var i int
        for {
                _, err := buf.Read(n)
                if err != nil {
                        break
                }
                if n[0] == 0x00 {
                        break
                }
                i++
        }
        return b[:i]
}