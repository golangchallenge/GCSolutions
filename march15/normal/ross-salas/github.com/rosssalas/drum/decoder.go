package drum

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"strings"
)

// Pattern contains high level splice data.
type Pattern struct {
	dataLen  int32   //length of data
	Hardware string  //hardware used to create file //[32]byte
	Tempo    float32 //tempo
	Tracks   []Track //slice of all tracks in file
}

// Track contains information for a single track.
type Track struct {
	ID      int32    //unique id
	nameLen byte     //length of track name
	Name    string   //track name
	Steps   [16]byte //notes for this track
}

// DecodeFile reads and decodes a file of type Splice.
// Input: use argument 'path' to specify the Splice filename
// Output: returns a *Pattern of the data read and decoded
// Returns error if any errors are encountered.
func DecodeFile(path string) (*Pattern, error) {
	f, err := os.Open(path)
	if err != nil {
		fmt.Println("error: DecodeFile() couldn't open file:", path)
		return nil, err
	}
	defer f.Close()
	p := &Pattern{}

	// read all header info
	err = p.readHeaders(f)
	if err != nil {
		fmt.Println("error: DecodeFile() error in file format:", err)
		return nil, err
	}

	// read all tracks
	err = p.readTracks(f)
	if err != nil {
		fmt.Println("error: DecodeFile() error in file format:", err)
		return nil, err
	}

	return p, nil
}

// readHeader reads all header info from splice file.
// It reads filetype marker, data lenth, hardware, and tempo.
func (p *Pattern) readHeaders(f *os.File) error {
	// read file type marker "SPLICE"
	var splice [10]byte
	err := binary.Read(f, binary.LittleEndian, &splice)
	if err != nil {
		fmt.Println("error: readHeader() read file type maker failed:", err)
		return err
	}
	if string(splice[:]) != "SPLICE\x00\x00\x00\x00" {
		err = errors.New("invalid file type marker")
		fmt.Println("error: readHeader()", err)
		return err
	}

	// read data length
	err = binary.Read(f, binary.BigEndian, &p.dataLen)
	if err != nil {
		fmt.Println("error: readHeader() read data length failed:", err)
		return err
	}
	p.dataLen = p.dataLen - 36

	// read hardware version
	var Hardware [32]byte //hardware is fixed length in Splice file and padded with 0's
	err = binary.Read(f, binary.LittleEndian, &Hardware)
	if err != nil {
		fmt.Println("error: readHeader() read hardware version failed:", err)
		return err
	}
	// hardware is easier to use in string format, convert to string then trim trailing 0's
	p.Hardware = strings.TrimRight(string(Hardware[:]), "\x00")

	// read tempo
	err = binary.Read(f, binary.LittleEndian, &p.Tempo)
	if err != nil {
		fmt.Println("error: readHeader() read tempo failed:", err)
		return err
	}

	return nil
}

// readTracks reads all tracks from splice file.
// p.dataLen specifies how much data needs to be read in.
func (p *Pattern) readTracks(f *os.File) error {
	// continue reading while p.dataLen is positive
	for p.dataLen > 0 {
		var t Track
		// read track id
		err := binary.Read(f, binary.LittleEndian, &t.ID)
		if err != nil {
			fmt.Println("error: readTrack() read track id failed:", err)
			return err
		}
		p.dataLen = p.dataLen - 4

		// read length of track name
		err = binary.Read(f, binary.LittleEndian, &t.nameLen)
		if err != nil {
			fmt.Println("error: readTrack() read track name length failed:", err)
			return err
		}
		p.dataLen = p.dataLen - 1

		// read track name
		Name := make([]byte, t.nameLen)
		err = binary.Read(f, binary.LittleEndian, &Name)
		if err != nil {
			fmt.Println("error: readTrack() read track name failed:", err)
			return err
		}
		t.Name = string(Name)
		p.dataLen = p.dataLen - int32(t.nameLen)

		// read steps  ie. binary [16]byte 1001100100001111
		err = binary.Read(f, binary.LittleEndian, &t.Steps)
		if err != nil {
			fmt.Println("error: readTrack() read steps failed:", err)
			return err
		}
		p.dataLen = p.dataLen - int32(16)

		// append this track to slice of tracks
		p.Tracks = append(p.Tracks, t)
	}

	return nil
}
