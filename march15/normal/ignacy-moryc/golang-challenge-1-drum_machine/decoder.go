package drum

import (
  "bytes"
  "encoding/binary"
  "fmt"
  "io/ioutil"
)

// track example:
// [...] 00 00 00 00 04 6b  69 63 6b 01 00 00 00 00
// 00 00 00 01 00 00 00 00  00 00 00
// First byte is Id (00) then there are 3 bytes (00 00 00) of padding
// then there's the name length (04) and after that the name (6b 69 63 6b)
// and what follows is 16 bytes describing the steps.
type track struct {
  ID    uint8
  Name  string
  Steps []byte
}

// patternHeader represents the header of a .splice file
//  53 50 4c 49 43 45 00 00  00 00 00 00 00 8f 30 2e  |SPLICE........0.|
//  38 30 38 2d 61 6c 70 68  61 00 00 00 00 00 00 00  |808-alpha.......|
//  00 00 00 00 00 00 00 00  00 00 00 00 00 00 cd cc  |................|
//  c4 42 [..]
type patternHeader struct {
  _       [14]byte // For 'SPLICE and the padding'
  Version [16]byte // Version signature
  _       [16]byte // Padding before Tempo and tracks
  Tempo   float32  // temopo
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
  Header patternHeader
  Tracks []track
}

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
  file, err := ioutil.ReadFile(path)
  if err != nil {
    fmt.Println(err)
    return nil, err
  }

  reader := &patternReader{bytes.NewReader(file)}

  ph := &patternHeader{}
  err = binary.Read(reader, binary.LittleEndian, ph)
  if err != nil {
    fmt.Println("reading header failed", err)
  }

  var tracks []track
  for reader.hasMorePotentialTracks() {
    track, err := reader.readNextTrack()
    if err == nil {
      tracks = append(tracks, *track)
    }
  }

  return &Pattern{Header: *ph, Tracks: tracks}, nil
}

func (p Pattern) String() string {
  var buffer bytes.Buffer
  buffer.WriteString(p.Header.String())
  for _, track := range p.Tracks[:] {
    buffer.WriteString(track.String())
  }
  return buffer.String()
}

func (h patternHeader) String() string {
  var buffer bytes.Buffer

  version := bytes.Trim(h.Version[:], "\x00")
  buffer.WriteString(fmt.Sprintf("Saved with HW Version: %s\n", version))

  // Float formatting:
  // - nothing after decimal point -> no decimal point (120 not 120.0)
  // - decimal part -> show 1 number after decimal point (92.5 not 93 or 92)
  if h.Tempo > float32(int(h.Tempo)) {
    buffer.WriteString(fmt.Sprintf("Tempo: %.1f\n", h.Tempo))
  } else {
    buffer.WriteString(fmt.Sprintf("Tempo: %d\n", int(h.Tempo)))
  }

  return buffer.String()
}

func (t track) String() string {
  return fmt.Sprintf("(%d) %s\t%s\n", t.ID, t.Name, t.stepsToString())
}

func (t *track) stepsToString() string {
  var buffer bytes.Buffer
  for i, el := range t.Steps[:] {
    if i%4 == 0 {
      buffer.WriteString("|")
    }

    if el == 1 {
      buffer.WriteString("x")
    } else {
      buffer.WriteString("-")
    }
  }

  buffer.WriteString("|")
  return buffer.String()
}
