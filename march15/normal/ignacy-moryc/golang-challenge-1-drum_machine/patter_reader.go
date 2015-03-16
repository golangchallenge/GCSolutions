package drum

import (
  "bytes"
  "encoding/binary"
  "errors"
)

const trackStepsCount = 16

// patternReader is a wrapper for bytes.Reader
type patternReader struct {
  *bytes.Reader
}

func (reader *patternReader) readNextTrack() (*track, error) {
  id, err := reader.readID()
  if err != nil {
    return nil, err
  }

  if reader.isLookingAtHeader(id) {
    return nil, errors.New("start of another pattern/repeated header")
  }

  // tracks have padding after id so we skip it:
  // move (3) bytes from current spot (1)
  reader.Seek(3, 1)

  name, err := reader.readName()
  if err != nil {
    return nil, err
  }

  steps, err := reader.readSteps()
  if err != nil {
    return nil, err
  }

  return &track{ID: id, Name: name, Steps: steps}, nil
}

func (reader *patternReader) readID() (byte, error) {
  return reader.ReadByte()
}

func (reader *patternReader) readName() (string, error) {
  nameLength, err := reader.ReadByte()
  if err != nil {
    return "", err
  }
  name := make([]byte, nameLength)
  err = binary.Read(reader, binary.LittleEndian, name)
  return string(name), err
}

func (reader *patternReader) readSteps() ([]byte, error) {
  steps := make([]byte, trackStepsCount)
  err := binary.Read(reader, binary.LittleEndian, steps)
  return steps[:], err
}

// Sometimes when one pattern ends another one starts, and we would
// like to stop reading tracks at this point. This function checks if
// buffer is looking at "SPLICE".  If not - reader gets rewinded back
// to the spot before checking.
func (reader *patternReader) isLookingAtHeader(id byte) bool {
  if string(id) == "S" {
    possibleSplice := make([]byte, 5)
    err := binary.Read(reader, binary.LittleEndian, possibleSplice)
    if err == nil && string(possibleSplice) == "PLICE" {
      for reader.Len() > 0 {
        _, err = reader.ReadByte()
      }
      return true
    }
    reader.Seek(-5, 1)
  }
  return false
}

func (reader *patternReader) hasMorePotentialTracks() bool {
  return reader.Len() > trackStepsCount
}
