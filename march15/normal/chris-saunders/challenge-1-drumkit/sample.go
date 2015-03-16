package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type sampleHeader struct {
	ID          int32
	LabelLength byte
}

// Sample includes beat details and a human readable label
type Sample struct {
	ID    int32
	Label string
	Steps [16]bool
}

// ExtractSample builds a sample from a stream
func ExtractSample(r io.Reader) (Sample, error) {
	var header sampleHeader
	sample := Sample{}
	err := binary.Read(r, binary.LittleEndian, &header)

	if err != nil {
		return sample, err
	}
	sample.ID = header.ID
	sample.Label, err = extractLabel(r, int(header.LabelLength))
	sample.Steps, err = extractSteps(r)
	return sample, err
}

func extractLabel(r io.Reader, size int) (string, error) {
	labelBytes := make([]byte, size)
	err := binary.Read(r, binary.LittleEndian, &labelBytes)
	if err != nil {
		return "", err
	}
	return string(labelBytes), nil
}

func extractSteps(r io.Reader) ([16]bool, error) {
	var steps [16]bool
	var rawSteps [16]byte
	err := binary.Read(r, binary.LittleEndian, &rawSteps)
	if err != nil {
		return steps, err
	}
	for i, b := range rawSteps {
		steps[i] = (b == 0x01)
	}
	return steps, nil
}

func (s Sample) String() string {
	return fmt.Sprintf("(%d) %s\t%s", s.ID, s.Label, s.stepsAsString())
}

func (s Sample) stepsAsString() string {
	buffer := bytes.NewBufferString("|")
	for i, b := range s.Steps {
		if b {
			buffer.WriteString("x")
		} else {
			buffer.WriteString("-")
		}
		if i%4 == 3 {
			buffer.WriteString("|")
		}
	}
	return buffer.String()
}
