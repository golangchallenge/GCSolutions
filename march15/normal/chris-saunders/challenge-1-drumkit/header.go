package drum

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

const (
	signatureLength int = 13
	versionLength   int = 32
)

type rawHeader struct {
	RawSignature [signatureLength]byte
	RawUnknown   byte
	RawVersion   [versionLength]byte
	RawTempo     float32
}

func (r rawHeader) version() string {
	for i := range r.RawVersion {
		if r.RawVersion[i] == 0x00 {
			return string(r.RawVersion[:i])
		}
	}
	return ""
}

func (r rawHeader) tempo() float64 {
	return float64(r.RawTempo)
}

const (
	floatingPointFormatStringForTempo string = "Saved with HW Version: %s\nTempo: %.1f"
	integerFormatStringForTempo       string = "Saved with HW Version: %s\nTempo: %.f"
)

// Header information contained within a splice file
type Header struct {
	Version string
	Tempo   float64
}

func (h Header) String() string {
	formatString := floatingPointFormatStringForTempo
	if math.Floor(h.Tempo) == h.Tempo {
		formatString = integerFormatStringForTempo
	}
	return fmt.Sprintf(formatString, h.Version, h.Tempo)
}

// ExtractHeader builds a Header from a stream
func ExtractHeader(r io.Reader) Header {
	var raw rawHeader
	err := binary.Read(r, binary.LittleEndian, &raw)
	if err != nil {
		fmt.Println("binary.Read failed:", err)
		return Header{}
	}
	return Header{Version: raw.version(), Tempo: raw.tempo()}
}
