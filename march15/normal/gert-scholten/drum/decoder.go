package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

const (
	spliceMagic = "SPLICE" // 6 bytes file magic
	versionSize = 32       // 32 bytes size of version string
)

// Decode decodes the drum machine data from a io.Reader
// and returns a pointer to a parsed pattern.
func Decode(r io.Reader) (*Pattern, error) {
	// File magic: "SPLICE"
	var magic [len(spliceMagic)]byte
	if _, err := io.ReadFull(r, magic[:]); err != nil {
		return nil, err
	}
	if !bytes.Equal(magic[:], []byte(spliceMagic)) {
		return nil, fmt.Errorf("file header not %v: %v", spliceMagic, string(magic[:]))
	}

	// 64bit size of bytes in the reader. Decremented while reading remaining data
	var size int64
	if err := binary.Read(r, binary.BigEndian, &size); err != nil {
		return nil, err
	}

	// 32 bytes version string
	var verbuf [versionSize]byte
	if _, err := io.ReadFull(r, verbuf[:]); err != nil {
		return nil, err
	}
	size -= versionSize
	// Version (can be/is) NUL terminated, or at least NUL padded
	verlen := bytes.IndexByte(verbuf[:], 0x00)
	if verlen < 0 {
		verlen = len(verbuf)
	}

	// Construct pattern
	p := &Pattern{}
	p.Version = string(verbuf[:verlen])
	if err := binary.Read(r, binary.LittleEndian, &p.Tempo); err != nil {
		return nil, err
	}
	size -= int64(binary.Size(&p.Tempo))

	// Read tracks
	for size > 0 {
		var t Track

		// 1 byte track id
		if err := binary.Read(r, binary.BigEndian, &t.ID); err != nil {
			return nil, err
		}
		size -= int64(binary.Size(&t.ID))

		// 4 bytes track name length
		var namelen uint32
		if err := binary.Read(r, binary.BigEndian, &namelen); err != nil {
			return nil, err
		}
		size -= int64(binary.Size(&namelen))

		// namelen bytes track name
		name := make([]byte, namelen)
		if _, err := io.ReadFull(r, name); err != nil {
			return nil, err
		}
		size -= int64(namelen)
		t.Name = string(name)

		// Steps bytes track data
		if _, err := io.ReadFull(r, t.Data[:]); err != nil {
			return nil, err
		}
		size -= int64(binary.Size(t.Data[:]))

		// Track read, append
		p.Tracks = append(p.Tracks, t)
	}
	if size != 0 {
		// more bytes were consumed than the size header field indicated
		return nil, fmt.Errorf("read too many bytes: %d", -size)
	}

	return p, nil
}

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern.
func DecodeFile(path string) (*Pattern, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return Decode(file)
}

func (p *Pattern) String() string {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, "Saved with HW Version: %s\n", p.Version)
	fmt.Fprintf(&buf, "Tempo: %v\n", p.Tempo)
	for _, t := range p.Tracks {
		fmt.Fprintf(&buf, "(%d) %s\t", t.ID, t.Name)
		for i, b := range t.Data {
			if i%4 == 0 {
				fmt.Fprint(&buf, "|")
			}
			if b > 0 {
				fmt.Fprint(&buf, "x")
			} else {
				fmt.Fprint(&buf, "-")
			}
		}
		fmt.Fprint(&buf, "|\n")
	}

	return buf.String()
}
