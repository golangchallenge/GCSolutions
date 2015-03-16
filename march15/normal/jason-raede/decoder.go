package drum

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"strings"
)

const versionStart = 14
const patchesStart = 50
const barLength = 16
const quarterNoteSeparator = "|"

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	p := &Pattern{}
	f, err := os.Open(path)

	if err != nil {
		return p, err
	}

	// We can also theoretically read the bytes one by one but the files are small enough that there's no memory concern and it is simpler to
	// have access to all of the data immediately since different properties live in different places.
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return p, err
	}

	// The tempo is defined by a four-byte slice from indeces 46-49 inclusive. This is a byte-representation of a float32.
	p.Tempo = floatFromByteSlice(data[46:50])

	p.Version = string(extractVersion(data))

	patches, err := getBeat(data)
	if err != nil {
		return p, err
	}
	p.Patches = patches

	return p, nil
}

// The version is an arbitrary string, terminated by a null-byte. So we keep going until we get a null byte.
func extractVersion(data []byte) []byte {
	version := []byte{}

	for _, b := range data[versionStart:] {
		if b == 0x00 {
			break
		}

		version = append(version, b)
	}

	return version
}

// To get the tempo as a float32 representation
func floatFromByteSlice(slc []byte) float32 {
	bits := binary.LittleEndian.Uint32(slc)
	float := math.Float32frombits(bits)
	return float
}

// Retrieve the "beat", which is the combination of one or more Patches. Each patch has a name, an identifier, and 16 notes (either "on" or "off")
//
// Patch byte structure is [identifier (4)] + [nameLength (1)] + [name (nameLength)] + [beat (16)]
func getBeat(data []byte) ([]*Patch, error) {
	var patch *Patch
	var err error

	p := []*Patch{}

	// The 14th byte tells the system how many more bytes to parse. For cases like pattern_5 there is MORE than the 87 bytes designated, but we won't even look at that.
	totalLength := int(data[13])

	// So the bytes we need to parse are from patchesStart until the end designated by byte 14. However the "end" as defined by byte 14 is FROM byte 14, so we need
	// to do some math here...
	remaining := data[patchesStart:(totalLength + 14)]

	// Keep extracting patches and resetting remaining to whatever is left. If everything goes to plan we should end up with exactly zero bytes remaining.
	for {
		patch, remaining, err = extractPatch(remaining)
		if err != nil {
			return p, err
		}
		p = append(p, patch)
		if len(remaining) == 0 {
			break
		} else if len(remaining) < (barLength + 5) {
			return p, errors.New("Invalid length for patch data")
		}
	}

	return p, nil

}

// Inspect the bytes in data to determine how many bytes comprise the next "patch definition". Then loop through those bytes and construct the patch
// by extracting the ID, name, and notes. Return the extracted patch, the remaining bytes that have not yet been parsed, and any error found.
func extractPatch(data []byte) (*Patch, []byte, error) {
	patch := &Patch{}

	// Indicator of length is at index 4
	patchNameLength := int(data[4])

	patch.Identifier = int(data[0])

	noteStart := patchNameLength + 5
	patch.Name = string(data[5:noteStart])

	notes := make([]bool, barLength)
	for i := noteStart; i < (noteStart + barLength); i++ {
		if data[i] == 0x00 {
			notes[i-noteStart] = false
		} else if data[i] == 0x01 {
			notes[i-noteStart] = true
		} else {
			return patch, []byte{}, errors.New("Invalid beat value (expected 0x00 or 0x01, got " + fmt.Sprint(data[i]))
		}
	}

	patch.Notes = notes
	return patch, data[(5 + barLength + patchNameLength):], nil
}

// Pattern is abstraction of the HW version, the tempo, and the patches w/ their notes.
// Patches could also be a map[string]*Patch where the key is the patch name or ID I suppose, but we need a way to keep them in order so a slice is probably best.
type Pattern struct {
	Version string
	Tempo   float32
	Patches []*Patch
}

func (p *Pattern) String() string {
	patchOutput := make([]string, len(p.Patches))

	for i, patch := range p.Patches {
		patchOutput[i] = patch.String()
	}

	// If it's an integer we want it to show as an integer, otherwise round to nearest 10th. So check that here and let the formatter do the work.
	tempoFormat := "%.0f"
	if math.Mod(float64(p.Tempo), 1.0) > 0 {
		tempoFormat = "%.1f"
	}
	return fmt.Sprintf("Saved with HW Version: %s\nTempo: "+tempoFormat+"\n%s%s", p.Version, p.Tempo, strings.Join(patchOutput, "\n"), "\n")
}

// SetNoteForPatchAtBeat ...
// Used to change the notes a patch plays after being extracted
// Note that this is the beat when counting as a musician, so the actual INDEX is (beat - 1)
func (p *Pattern) SetNoteForPatchAtBeat(patchID int, beat int, note bool) error {
	if beat >= 16 {
		return errors.New("There are only 16 beats in a measure with this drum machine!")
	}

	if beat == 0 {
		return errors.New("There is no beat 0!")
	}

	// First get the patch - if not found we'll return an error.
	for _, patch := range p.Patches {
		if patch.Identifier == patchID {
			patch.Notes[beat-1] = note
			return nil
		}
	}

	return errors.New("Patch not found")
}

// Patch ...
// Each patch is represented by a slice of 16 "Sixteenth Notes", which are either "on" (true) or "off" (false).
type Patch struct {
	Name       string
	Notes      []bool
	Identifier int
}

func (p *Patch) String() string {
	beat := ""

	// Every 4th note we want to add a |, including before the first note.
	// If the note is true ("on/played") it's an "x". Otherwise a "-"
	for i := 0; i < barLength; i++ {
		if math.Mod(float64(i), 4.0) == 0 {
			beat += quarterNoteSeparator
		}
		if p.Notes[i] {
			beat += "x"
		} else {
			beat += "-"
		}

	}
	beat += quarterNoteSeparator

	return fmt.Sprintf("(%d) %s\t%s", p.Identifier, p.Name, beat)
}
